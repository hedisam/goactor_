package context

import (
	"github.com/hedisam/goactor/internal/mailbox"
	"time"
)

type Context interface {
	Args() []interface{}
	Recv(mailbox.MessageHandler)
	RecvWithTimeout(time.Duration, mailbox.MessageHandler)

	// Done() returns a channel that can be used to know if the actor is been shutdown or not.
	// users should listen for the channel in case of long running tasks, if closed, terminate by returning.
	// todo: terminate by a specific message?
	Done() <-chan struct{}
}
