package actor

import (
	"github.com/hedisam/goactor/internal/context"
	"github.com/hedisam/goactor/internal/pid"
)

const (
	trapExitNo int32 = iota
	trapExitYes
)

const (
	// actor types
	WorkerActor int32 = iota
	SupervisorActor
)

type Func func(actor Actor)

type Actor interface {
	TrapExit(trap bool)
	Monitor(pid *pid.ProtectedPID)
	Demonitor(pid *pid.ProtectedPID)
	Link(pid *pid.ProtectedPID)
	Unlink(pid *pid.ProtectedPID)
	SpawnLink(fn Func, args ...interface{}) *pid.ProtectedPID
	SpawnMonitor(fn Func, args ...interface{}) *pid.ProtectedPID
	Context() context.Context
	Self() *pid.ProtectedPID
	handleTermination()
	setActorType(_type int32)
	actorType() int32
}

