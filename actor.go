package goactor

import (
	"fmt"
	"sync/atomic"
)

const (
	actorTrapExitNo int32 = iota
	actorTrapExitYes
)

type ActorFunc func(actor *Actor)

type Actor struct {
	*Context
	trapExit   		int32
	// actors that are linked to me. two way communication
	linkedActors 	map[string]*Actor
	// actors that are monitoring me. one way communication
	monitorActors	map[string]*Actor
}

func newActor(ctx *Context) *Actor {
	return &Actor{
		Context:       ctx,
		trapExit:      actorTrapExitNo,
		linkedActors:  make(map[string]*Actor),
		monitorActors: make(map[string]*Actor),
	}
}

func (actor *Actor) linkTo(linkedActor *Actor) *Actor {
	actor.linkedActors[linkedActor.pid.id] = linkedActor
	return actor
}

func (actor *Actor) unlinkFrom(linkedActor *Actor) *Actor {
	delete(actor.linkedActors, linkedActor.pid.id)
	return actor
}

func (actor *Actor) monitoredBy(monitorActor *Actor) *Actor {
	actor.monitorActors[monitorActor.pid.id] = monitorActor
	return actor
}

func NewParentActor() *Actor {
	// todo: this is not a panic-safe actor, fix it
	pid := newPID()
	actor := newActor(newContext(pid))
	pid.mailbox.setActor(actor)
	return actor
}

func (actor *Actor) TrapExit(trapExit bool) {
	var trap = actorTrapExitNo
	if trapExit {
		trap = actorTrapExitYes
	}
	atomic.StoreInt32(&actor.trapExit, trap)
}

func (actor *Actor) Monitor(pid *PID) {
	request := MonitorRequest{
		who:       pid.mailbox.getActor(),
		by:        actor,
		demonitor: false,
	}
	sendSystem(pid, request)

}

func (actor *Actor) DeMonitor(pid *PID) {
	request := MonitorRequest{
		who:       pid.mailbox.getActor(),
		by:        actor,
		demonitor: true,
	}
	sendSystem(pid, request)
}

func (actor *Actor) Link(pid *PID) {
	request := LinkRequest{who: pid.mailbox.getActor(), to: actor, unlink: false}
	sendSystem(pid, request)
	actor.linkTo(pid.mailbox.getActor())
}

func (actor *Actor) Unlink(pid *PID) {
	request := LinkRequest{who: pid.mailbox.getActor(), to: actor, unlink: true}
	sendSystem(pid, request)
	actor.unlinkFrom(pid.mailbox.getActor())
}

// SpawnLink spawns a new actor linked to to the caller actor
func (actor *Actor) SpawnLink(fn ActorFunc, args ...interface{}) *PID {
	pid := newPID()
	ctx := newContext(pid).withActorFunc(fn).withArgs(args)
	linkedActor := newActor(ctx).linkTo(actor)
	pid.mailbox.setActor(linkedActor)
	actor.linkTo(linkedActor)
	spawn(linkedActor)
	return pid
}

// SpawnMonitor spawns a new actor monitored by the caller actor
func (actor *Actor) SpawnMonitor(fn ActorFunc, args ...interface{}) *PID {
	pid := newPID()
	ctx := newContext(pid).withActorFunc(fn).withArgs(args)
	monitoredActor := newActor(ctx).monitoredBy(actor)
	pid.mailbox.setActor(monitoredActor)
	spawn(monitoredActor)
	return pid
}

// Spawn spawns a function as an actor also passing its args to the actor
func Spawn(fn ActorFunc, args ...interface{}) *PID {
	pid := newPID()
	ctx := newContext(pid).withActorFunc(fn).withArgs(args)
	actor := newActor(ctx)
	pid.mailbox.setActor(actor)
	spawn(actor)
	return pid
}

// Send a message to an actor by its pid
func Send(pid *PID, message interface{}) {
	pid.mailbox.sendUserMessage(message)
}

func sendSystem(pid *PID, message SystemMessage) {
	pid.mailbox.sendSysMessage(message)
}

func spawn(actor *Actor) {
	go func(actor *Actor) {
		// handleTermination args and receiver will be evaluated when defer schedules our method call not when
		// the actual method executes. other actors could be added or removed from the linked/monitored/monitors
		// list while the actor is running but we'll not be worry about it since our receiver (actor) is a pointer.
		defer actor.handleTermination()
		actor.fn(actor)
	}(actor)
}

func (actor *Actor) handleTermination() {
	// close actor's close channel so it doesn't accept any further messages
	//actor.pid.mailbox.close()

	// check if we got a panic or just a normal termination
	switch r := recover().(type) {
	case PauseCMD:
		// we're gonna pause the actor. how? just by returning. notice that we're not closing the mailbox.
		//log.Println("[-] actor paused, id:", actor.pid.id)
		return
	case ExitCMD:
		actor.pid.mailbox.close()
		// got an exit command. it means one of the linked actors had a panic
		// let's notify our monitor actors
		killExit := KillExit{who: actor, by: r.becauseOf, reason: r.reason}
		actor.notifyMonitors(killExit)
		// notify our linked actors, except the one causing this.
		actor.notifyLinkedActors(killExit)
	default:
		actor.pid.mailbox.close()
		if r != nil {
			// we're caused the the panic
			// todo: use better reasons. reason's type should be interface{}
			reason := fmt.Sprint(r)
			actor.notifyMonitors(PanicExit{who: actor, reason: reason})
			actor.notifyLinkedActors(ExitCMD{becauseOf: actor, reason: reason})
		} else {
			// it's just a normal termination
			normalExit := NormalExit{who: actor}
			actor.notifyMonitors(normalExit)
			actor.notifyLinkedActors(normalExit)
		}
	}
}

func (actor *Actor) notifyMonitors(message SystemMessage) {
	for _, monitor := range actor.monitorActors {
		sendSystem(monitor.pid, message)
	}
}

func (actor *Actor) notifyLinkedActors(message SystemMessage) {
	for _, linked := range actor.linkedActors {
		switch msg := message.(type) {
		case ExitCMD, NormalExit:
			sendSystem(linked.pid, message)
		case KillExit:
			if msg.by.pid.id == linked.pid.id {
				// it's termination's source. so we don't need to notify
				continue
			}
			sendSystem(linked.pid, ExitCMD{becauseOf: actor, reason: msg.reason})
		}
	}
}

