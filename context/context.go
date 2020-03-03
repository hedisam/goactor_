package context

import (
	"github.com/hedisam/goactor/internal/mailbox"
	"time"
)

type Context interface {
	Args() []interface{}
	Recv(mailbox.MessageHandler)
	RecvWithTimeout(time.Duration, mailbox.MessageHandler)
	Done() <-chan struct{}
}
