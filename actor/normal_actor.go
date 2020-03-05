package actor

import (
	"github.com/hedisam/goactor/context"
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
}

func newActor(ctx context.Context, utils *mailbox.ActorUtils) *actor {
	actor := &actor{
		context:       ctx,
		trapExit:      trapExitNo,
		linkedActors:  make(map[pid.PID]pid.PID),
		monitorActors: make(map[pid.PID]pid.PID),
	}
	actor.setUtils(utils)
	return actor
}

func (a *actor) withPID(_pid pid.PID) *actor {
	a.self = pid.NewProtectedPID(_pid)
	return a
}

func (a *actor) build() Actor {
	return a
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
		a.notifyLinkedActors(r)
		r1 := r
		r1.Relation = sysmsg.Monitored
		a.notifyLinkedActors(r)
	default:
		// something went wrong. notify monitors and linked actors with an Exit msg
		if r != nil {
			//log.Println(r)
			exit := sysmsg.Exit{
				Who:      pid.ExtractPID(a.self),
				Reason:   sysmsg.Panic,
				Relation: sysmsg.Linked,
			}
			a.notifyLinkedActors(exit)
			exit1 := exit
			exit1.Relation = sysmsg.Monitored
			a.notifyMonitors(exit)

			// it's a normal exit
		} else {
			normal := sysmsg.Exit{
				Who:      pid.ExtractPID(a.self),
				Reason:   sysmsg.Normal,
				Relation: sysmsg.Linked,
			}
			a.notifyLinkedActors(normal)
			normal1 := normal
			normal1.Relation = sysmsg.Monitored
			a.notifyMonitors(normal1)
		}
	}
}

func (a *actor) notifyMonitors(message sysmsg.Exit) {
	for _, monitor := range a.monitorActors {
		sendSystemMessage(pid.NewProtectedPID(monitor), message)
	}
}

func (a *actor) notifyLinkedActors(message sysmsg.Exit) {
	for _, linked := range a.linkedActors {
		if linked != message.Parent {
			// todo: what if the linked parent actor is a supervisor?
			sendSystemMessage(pid.NewProtectedPID(linked), message)
		}
	}
}

func fromInterface(p interface{}) (_pid pid.PID) {
	_pid, _ = p.(pid.PID)
	return
}