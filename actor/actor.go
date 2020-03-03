package actor

import (
	"github.com/hedisam/goactor/context"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/internal/sysmsg"
)

const (
	trapExitNo int32 = iota
	trapExitYes
)

type Func func(actor Actor)

type Actor interface {
	TrapExit(trap bool)
	Monitor(pid *PID)
	Demonitor(pid *PID)
	Link(pid *PID)
	Unlink(pid *PID)
	SpawnLink(fn Func, args ...interface{}) *PID
	SpawnMonitor(fn Func, args ...interface{}) *PID
	Context() context.Context
	Self() *PID
	handleTermination()
}

type PID struct {
	pid pid.PID
}

func FromInterface(p interface{}) pid.PID {
	return p.(pid.PID)
}

func Send(pid *PID, message interface{}) {
	pid.pid.Mailbox().SendUserMessage(message)
}

func SendNamed(name string, message interface{}) {
	//_pid := process.WhereIs(name)
	//if _pid == nil {
	//	return
	//}
	//Send(_pid, message)
}

func Spawn(fn Func, args ...interface{}) *PID {
	utils := &mailbox.ActorUtils{}
	_pid := pid.NewPID(utils)
	ctx := context.NewContext(_pid, args)
	actor := newActor(ctx, utils).withPID(_pid)
	spawn(fn, actor)
	return &PID{pid: _pid}
}

func spawn(fn Func, actor Actor) {
	go func(fn Func, actor Actor) {
		defer actor.handleTermination()
		fn(actor)
	}(fn, actor)
}

func sendSystemMessage(pid *PID, message sysmsg.SystemMessage) {
	pid.pid.Mailbox().SendSystemMessage(message)
}
