package actor

import (
	"github.com/hedisam/goactor/internal/context"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/sysmsg"
)

func Send(ppid *pid.ProtectedPID, message interface{}) {
	pid.ExtractPID(ppid).Mailbox().SendUserMessage(message)
}

func SendNamed(name string, message interface{}) {
	ppid := WhereIs(name)
	if ppid == nil {return}
	Send(ppid, message)
}

func Spawn(fn Func, args ...interface{}) *pid.ProtectedPID {
	actor := createActor(args...)
	spawn(fn, actor)
	return actor.Self()
}

func spawnLink(fn Func, to pid.PID, args ...interface{}) *pid.ProtectedPID {
	actor := createActor(args...)
	actor.link(to)
	spawn(fn, actor)
	return actor.Self()
}

func spawnMonitor(fn Func, by pid.PID, args ...interface{}) *pid.ProtectedPID {
	actor := createActor(args...)
	actor.monitoredBy(by)
	spawn(fn, actor)
	return actor.Self()
}

func createActor(args ...interface{}) *Actor {
	utils := &mailbox.ActorUtils{}
	_pid := pid.NewPID(utils)
	ctx := context.NewContext(_pid, args)
	actor := newActor(ctx, _pid, utils)
	return actor
}

func spawn(fn Func, actor *Actor) {
	go func() {
		defer actor.handleTermination()
		fn(actor)
	}()
}

func sendSystemMessage(ppid *pid.ProtectedPID, message sysmsg.SystemMessage) {
	pid.ExtractPID(ppid).Mailbox().SendSystemMessage(message)
}
