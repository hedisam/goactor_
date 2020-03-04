package actor

import (
	"github.com/hedisam/goactor/context"
	"github.com/hedisam/goactor/internal/pid"
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

