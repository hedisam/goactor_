package goactor

import "github.com/rs/xid"

type PID struct {
	mailbox Mailbox
	id      string
}

func (pid *PID) ID() string {
	return pid.id
}

func newPID() *PID {
	return &PID{
		mailbox: defaultQueueMailbox(),
		id:      xid.New().String(),
	}
}
