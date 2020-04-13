package pid

import (
	"context"
	"github.com/rs/xid"
)

type MailboxWriter interface {
	SendUserMessage(message interface{})
	SendSystemMessage(message interface{})
	Dispose()
}

type BasePID struct {
	MailboxWriter
	id string
}

func NewFuturePID(mailbox MailboxWriter) *BasePID {
	return &BasePID{
		MailboxWriter: mailbox,
		id:            xid.New().String(),
	}
}

func (pid *BasePID) ID() string {
	return pid.id
}

type PID struct {
	BasePID
	cancel context.CancelFunc
}

func NewPID(mailbox MailboxWriter, cancel context.CancelFunc) *PID {
	return &PID{
		BasePID: *NewFuturePID(mailbox),
		cancel:  cancel,
	}
}

// Shutdown closes the done channel of actor's context.Context object by calling its cancel function
// used by supervisor
func (pid *PID) Shutdown() {
	pid.cancel()
}