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
	utils := &mailbox.ActorUtils{}
	_pid := pid.NewPID(utils)
	ctx := context.NewContext(_pid, args)
	actor := newActor(ctx, utils).withPID(_pid)
	spawn(fn, actor)
	return pid.NewProtectedPID(_pid)
}

func spawn(fn Func, actor Actor) {
	go func(fn Func, actor Actor) {
		defer actor.handleTermination()
		fn(actor)
	}(fn, actor)
}

func sendSystemMessage(ppid *pid.ProtectedPID, message sysmsg.SystemMessage) {
	pid.ExtractPID(ppid).Mailbox().SendSystemMessage(message)
}
