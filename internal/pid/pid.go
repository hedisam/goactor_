package pid

import (
	"github.com/hedisam/goactor/internal/mailbox"
)

type PID interface {
	Mailbox() mailbox.Mailbox

	// ShutdownFn() returns a function that can be used to close the context's done channel.
	// used by supervisor when shutting down an actor
	ShutdownFn() func()
	SetShutdownFn(fn func())

	// default actor type is "child", but it could be a supervisor too.
	SetActorTypeFn(fn func(int32))
	ActorTypeFn() func(int32)
}

type ProtectedPID struct {
	pid PID
}

func NewProtectedPID(pid PID) *ProtectedPID {
	return &ProtectedPID{pid: pid}
}

func ExtractPID(ppid *ProtectedPID) PID {
	return ppid.pid
}