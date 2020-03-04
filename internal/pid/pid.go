package pid

import (
	"github.com/hedisam/goactor/internal/mailbox"
)

type PID interface {
	Mailbox() mailbox.Mailbox
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