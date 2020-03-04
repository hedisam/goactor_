package pid

import (
	"github.com/hedisam/goactor/internal/mailbox"
)

type PID interface {
	Mailbox() mailbox.Mailbox

	// Shutdown() returns a function that can be used to close the context's done channel.
	// used by supervisor when shutting down an actor
	Shutdown() func()
	SetShutdown(fn func())
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