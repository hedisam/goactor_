package pid

import (
	"github.com/hedisam/goactor/internal/mailbox"
)

type PID interface {
	Mailbox() mailbox.Mailbox
}

type localPID struct {
	m mailbox.Mailbox
}

func NewPID(utils *mailbox.ActorUtils) PID {
	return &localPID{
		mailbox.DefaultRingBufferQueueMailbox(utils),
	}
}

func (pid *localPID) Mailbox() mailbox.Mailbox {
	return pid.m
}
