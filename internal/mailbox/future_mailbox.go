package mailbox

import (
	"fmt"
	"github.com/hedisam/goactor/sysmsg"
	"time"
)

var ErrDisposed = fmt.Errorf("mailbox's channel is closed")

type FutureMailbox struct {
	m chan interface{}
	done chan struct{}
}

func NewFutureMailbox() *FutureMailbox {
	return &FutureMailbox{
		m:    make(chan interface{}, 1),
		done: make(chan struct{}),
	}
}

func (f *FutureMailbox) SendUserMessage(message interface{}) {
	select {
	case <-f.done:
	case f.m<- message:
	}
}

func (f *FutureMailbox) SendSystemMessage(message interface{}) {
	f.SendUserMessage(message)
}

func (f *FutureMailbox) Receive(handler func(message interface{}) (loop bool)) {
	select {
	case msg := <-f.m:
		handler(msg)
	case <-f.done:
		handler(ErrDisposed)
	}
}

func (f *FutureMailbox) ReceiveWithTimeout(d time.Duration, handler func(message interface{}) (loop bool)) {
	select {
	case msg := <-f.m:
		handler(msg)
	case <-time.After(d):
		handler(sysmsg.Timeout{})
	case <-f.done:
		handler(ErrDisposed)
	}
}

func (f *FutureMailbox) Dispose() {
	close(f.done)
}

