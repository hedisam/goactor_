package actor

import (
	"github.com/hedisam/goactor/internal/context"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/pid"
	"log"
)

type Func func(actor *Actor)

func Send(pid UserPID, message interface{}) {
	pid.SendUserMessage(message)
}

func SendNamed(name string, message interface{}) {
	namedPID := registry.WhereIs(name)
	if namedPID == nil {
		log.Println("SendNamed, pid not found:", name)
		return
	}
	Send(namedPID, message)
}

func Spawn(fn Func, args ...interface{}) UserPID {
	actor := createActor(args...)
	spawn(fn, actor)
	return actor.Self()
}

func spawnLink(fn Func, to UserPID, args ...interface{}) UserPID {
	actor := createActor(args...)
	actor.connectedActors.link(to)
	spawn(fn, actor)
	return actor.Self()
}

func spawnMonitor(fn Func, by UserPID, args ...interface{}) UserPID {
	actor := createActor(args...)
	actor.connectedActors.monitoredBy(by)
	spawn(fn, actor)
	return actor.Self()
}

func createActor(args ...interface{}) *Actor {
	m := mailbox.DefaultRingBufferQueueMailbox()
	ctx, cancelFunc := context.NewContext(m, args)
	actorPID := pid.NewPID(m, cancelFunc)
	actor := newActor(ctx, actorPID)
	m.SetSystemMessageHandler(&systemHandler{actor: actor})
	return actor
}

func spawn(fn Func, actor *Actor) {
	go func() {
		defer actor.handleTermination()
		fn(actor)
	}()
}
