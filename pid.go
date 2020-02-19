package goactor

import "github.com/rs/xid"

type PID struct {
	mailbox *mailbox
	id 	string
}

func (pid *PID) ID() string {
	return pid.id
}

func newPID() *PID {
	return &PID{
		mailbox: defaultMailbox(),
		id: xid.New().String(),
	}
}
