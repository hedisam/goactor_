package actor

import (
	"github.com/hedisam/goactor/internal/context"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/sysmsg"
	"sync/atomic"
)

type actor struct {
	context  context.Context
	trapExit int32
	// actors that are linked to me. two way communication
	linkedActors map[pid.PID]pid.PID
	// actors that are monitoring me. one way communication
	monitorActors map[pid.PID]pid.PID
	self          *pid.ProtectedPID
	// actor type: WorkerActor or SupervisorActor
	aType	int32
	supervisedBy	pid.PID
}

func newActor(ctx context.Context, _pid pid.PID , utils *mailbox.ActorUtils) Actor {
	actor := &actor{
		context:       ctx,
		trapExit:      trapExitNo,
		linkedActors:  make(map[pid.PID]pid.PID),
		monitorActors: make(map[pid.PID]pid.PID),
		self:          pid.NewProtectedPID(_pid),
		aType:         WorkerActor,
	}
	actor.setUtils(utils)
	_pid.SetActorTypeFn(actor.setActorType)
	_pid.SetSupervisorFn(actor.setSupervisor)
	return actor
}

// setSupervisor must only be called once, right after spawning by a supervisor
func (a *actor) setSupervisor(_pid pid.PID) {
	a.supervisedBy = _pid
}

// supervisor only needed when handling termination, notifying linked actors
func (a *actor) supervisor() pid.PID {
	return a.supervisedBy
}

// setActorType sets actor type to WorkerActor == 0 or SupervisorActor == 1
// default is WorkerActor
func (a *actor) setActorType(_type int32) {
	atomic.StoreInt32(&a.aType, _type)
}

func (a *actor) actorType() int32 {
	return atomic.LoadInt32(&a.aType)
}

func (a *actor) setUtils(utils *mailbox.ActorUtils) {
	utils.Link = func(pid interface{}) {
		a.link(fromInterface(pid))
	}
	utils.Unlink = func(pid interface{}) {
		a.unlink(fromInterface(pid))
	}
	utils.MonitoredBy = func(pid interface{}) {
		a.monitoredBy(fromInterface(pid))
	}
	utils.DemonitorBy = func(pid interface{}) {
		a.demoniteredBy(fromInterface(pid))
	}
	utils.Self = func() interface{} {
		return pid.ExtractPID(a.Self())
	}
	utils.ContextDone = func() <-chan struct{} {
		return a.Context().Done()
	}
	utils.TrapExit = a.trapExited
}

// util methods. these methods must only be called from
// mailbox receive when handling system messages

func (a *actor) link(pid pid.PID) {
	a.linkedActors[pid] = pid
}

func (a *actor) unlink(pid pid.PID) {
	delete(a.linkedActors, pid)
}

func (a *actor) monitoredBy(pid pid.PID) {
	a.monitorActors[pid] = pid
}

func (a *actor) demoniteredBy(pid pid.PID) {
	delete(a.monitorActors, pid)
}

func (a *actor) trapExited() (trap bool) {
	trap = false
	if atomic.LoadInt32(&a.trapExit) == trapExitYes {
		trap = true
	}
	return
}

func (a *actor) Monitor(ppid *pid.ProtectedPID) {
	request := sysmsg.Monitor{Parent: pid.ExtractPID(a.self)}
	sendSystemMessage(ppid, request)
}

func (a *actor) Demonitor(ppid *pid.ProtectedPID) {
	request := sysmsg.Monitor{Parent: pid.ExtractPID(a.self), Revert: true}
	sendSystemMessage(ppid, request)
}

func (a *actor) Link(ppid *pid.ProtectedPID) {
	// send a link request to the target actor
	req := sysmsg.Link{To: pid.ExtractPID(a.self)}
	sendSystemMessage(ppid, req)
	// send a link request to ourselves
	req1 := sysmsg.Link{To: pid.ExtractPID(ppid)}
	sendSystemMessage(a.Self(), req1)
}

func (a *actor) Unlink(ppid *pid.ProtectedPID) {
	// send an unlink request to the target actor
	req := sysmsg.Link{To: pid.ExtractPID(a.self), Revert: true}
	sendSystemMessage(ppid, req)
	// send an unlink request to ourselves
	req1 := sysmsg.Link{To: pid.ExtractPID(ppid), Revert: true}
	sendSystemMessage(a.Self(), req1)
}

func (a *actor) SpawnLink(fn Func, args ...interface{}) *pid.ProtectedPID {
	ppid := Spawn(fn, args...)
	a.Link(ppid)
	return ppid
}

func (a *actor) SpawnMonitor(fn Func, args ...interface{}) *pid.ProtectedPID {
	ppid := Spawn(fn, args...)
	a.Monitor(ppid)
	return ppid
}

func (a *actor) TrapExit(trapExit bool) {
	var trap = trapExitNo
	if trapExit {
		trap = trapExitYes
	}
	atomic.StoreInt32(&a.trapExit, trap)
}

func (a *actor) Context() context.Context {
	return a.context
}

func (a *actor) Self() *pid.ProtectedPID {
	return a.self
}

func (a *actor) handleTermination() {
	// close actor's mailbox done channel so it can't accept any further messages
	pid.ExtractPID(a.self).Mailbox().Dispose()

	// check if we got a panic or just a normal return
	switch r := recover().(type) {

	// a linked actor terminated or got a sysmsg.Shutdown command by a supervisor
	// notify monitors and other linked actors
	case sysmsg.Exit:
		a.notifyLinkedActors(r, false)
		a.notifyMonitors(r)
	case sysmsg.Shutdown:
		// there's a case where user trap exit and receives the sysmsg.Shutdown msg then returns with same msg
		exit := sysmsg.Exit{
			Who:      a.self,
			Parent:   r.Parent,
			Reason:   sysmsg.Kill,
		}
		a.notifyLinkedActors(exit, false)
		a.notifyMonitors(exit)
	default:
		// something went wrong. notify monitors and linked actors with an Exit msg
		if r != nil {
			//log.Println(r)
			exit := sysmsg.Exit{
				Who:      pid.ExtractPID(a.self),
				Reason:   sysmsg.Panic,
			}

			// check if the actor is a supervisor, if so, then it had an unexpected panic and got no chances to
			// shutdown its children
			shutdownChildren := a.actorType() == SupervisorActor
			a.notifyLinkedActors(exit, shutdownChildren)
			a.notifyMonitors(exit)

			// it's a normal exit
		} else {
			normal := sysmsg.Exit{
				Who:      pid.ExtractPID(a.self),
				Reason:   sysmsg.Normal,
			}
			a.notifyLinkedActors(normal, false)
			a.notifyMonitors(normal)
		}
	}
}

func (a *actor) notifyMonitors(message sysmsg.Exit) {
	message.Relation = sysmsg.Monitored
	for _, monitor := range a.monitorActors {
		sendSystemMessage(pid.NewProtectedPID(monitor), message)
	}
}

func (a *actor) notifyLinkedActors(message sysmsg.Exit, shutdown bool) {
	message.Relation = sysmsg.Linked
	for _, linked := range a.linkedActors {
		sendSystemMessage(pid.NewProtectedPID(linked), message)
		// we can't shutdown our parent supervisor
		if shutdown && a.supervisor() != linked {
			linked.ShutdownFn()()
		}
	}
}

func fromInterface(p interface{}) (_pid pid.PID) {
	_pid, _ = p.(pid.PID)
	return
}