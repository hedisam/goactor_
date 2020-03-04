package mailbox

import (
	"github.com/hedisam/goactor/sysmsg"
	"time"
)

type ErrDisposed string

type future struct {
	m chan interface{}
	done chan struct{}
}

func NewFutureMailbox() *future {
	return &future{
		m:    make(chan interface{}, 1),
		done: make(chan struct{}),
	}
}

func (f *future) SendUserMessage(message interface{}) {
	select {
	case <-f.done:
	case f.m<- message:
	}
}

func (f *future) SendSystemMessage(message interface{}) {
	f.SendUserMessage(message)
}

func (f *future) Receive(handler MessageHandler) {
	select {
	case msg := <-f.m:
		handler(msg)
	case <-f.done:
		handler(ErrDisposed("mailbox's channel is closed"))
	}
}

func (f *future) ReceiveWithTimeout(d time.Duration, handler MessageHandler) {
	select {
	case msg := <-f.m:
		handler(msg)
	case <-time.After(d):
		handler(sysmsg.Timeout{})
	case <-f.done:
		handler(ErrDisposed("mailbox's channel is closed"))
	}
}

func (f *future) Dispose() {
	close(f.done)
}

// Utils returns nil. DO NOT call me
func (f *future) Utils() *ActorUtils {
	return nil
}

