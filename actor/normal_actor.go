package actor

import (
	"github.com/hedisam/goactor/context"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/internal/sysmsg"
	"sync/atomic"
)

type actor struct {
	context  context.Context
	trapExit int32
	// actors that are linked to me. two way communication
	linkedActors map[pid.PID]pid.PID
	// actors that are monitoring me. one way communication
	monitorActors map[pid.PID]pid.PID
	self          *PID
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

func (a *actor) withPID(pid pid.PID) *actor {
	a.self = &PID{pid}
	return a
}

func (a *actor) build() Actor {
	return a
}

func (a *actor) setUtils(utils *mailbox.ActorUtils) {
	utils.Link = func(pid interface{}) {
		a.link(FromInterface(pid))
	}
	utils.Unlink = func(pid interface{}) {
		a.unlink(FromInterface(pid))
	}
	utils.MonitoredBy = func(pid interface{}) {
		a.monitoredBy(FromInterface(pid))
	}
	utils.DemonitorBy = func(pid interface{}) {
		a.demoniteredBy(FromInterface(pid))
	}
	utils.Self = func() interface{} {
		return a.Self().pid
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

func (a *actor) Monitor(pid *PID) {
	request := sysmsg.Monitor{Parent: a.Self().pid}
	sendSystemMessage(pid, request)
}

func (a *actor) Demonitor(pid *PID) {
	request := sysmsg.Monitor{Parent: a.Self().pid, Revert: true}
	sendSystemMessage(pid, request)
}

func (a *actor) Link(pid *PID) {
	// send a link request to the target actor
	req := sysmsg.Link{To: a.Self().pid}
	sendSystemMessage(pid, req)
	// send a link request to ourselves
	req1 := sysmsg.Link{To: pid.pid}
	sendSystemMessage(a.Self(), req1)
}

func (a *actor) Unlink(pid *PID) {
	// send an unlink request to the target actor
	req := sysmsg.Link{To: a.Self().pid, Revert: true}
	sendSystemMessage(pid, req)
	// send an unlink request to ourselves
	req1 := sysmsg.Link{To: pid.pid, Revert: true}
	sendSystemMessage(a.Self(), req1)
}

func (a *actor) SpawnLink(fn Func, args ...interface{}) *PID {
	_pid := Spawn(fn, args)
	a.Link(_pid)
	return _pid
}

func (a *actor) SpawnMonitor(fn Func, args ...interface{}) *PID {
	_pid := Spawn(fn, args)
	a.Monitor(_pid)
	return _pid
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

func (a *actor) Self() *PID {
	return a.self
}

func (a *actor) handleTermination() {
	// close actor's mailbox done channel so it can't accept any further messages
	a.Self().pid.Mailbox().Dispose()

	// check if we got a panic or just a normal return
	switch r := recover().(type) {

	// a linked actor terminated
	// notify monitors and other linked actors
	case sysmsg.Exit:
		a.notifyLinkedActors(r)
		r1 := r
		r1.Relation = sysmsg.Monitored
		a.notifyLinkedActors(r)

	// someone commanded us to shutdown.
	case sysmsg.Shutdown:
		exit := sysmsg.Exit{
			Who:      a.Self().pid,
			Parent:   r.Parent,
			Reason:   sysmsg.Kill,
			Relation: sysmsg.Linked,
		}
		a.notifyLinkedActors(exit)
		exit1 := exit
		exit1.Relation = sysmsg.Monitored
		a.notifyMonitors(exit1)

	default:
		// something went wrong. notify monitors and linked actors with an Exit msg
		if r != nil {
			exit := sysmsg.Exit{
				Who:      a.Self().pid,
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
				Who:      a.Self().pid,
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
		sendSystemMessage(&PID{monitor}, message)
	}
}

func (a *actor) notifyLinkedActors(message sysmsg.Exit) {
	for _, linked := range a.linkedActors {
		if linked != message.Parent {
			// todo: what if the linked parent actor is a supervisor?
			sendSystemMessage(&PID{linked}, message)
		}
	}
}
