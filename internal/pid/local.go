package pid

import "github.com/hedisam/goactor/internal/mailbox"

type localPID struct {
	m mailbox.Mailbox
	shutdown func()
}

func NewPID(utils *mailbox.ActorUtils) PID {
	return &localPID{
		m: mailbox.DefaultRingBufferQueueMailbox(utils),
	}
}

func (pid *localPID) Mailbox() mailbox.Mailbox {
	return pid.m
}

func (pid *localPID) Shutdown() func() {
	return pid.shutdown
}

func (pid *localPID) SetShutdown(shutdown func()) {
	pid.shutdown = shutdown
}