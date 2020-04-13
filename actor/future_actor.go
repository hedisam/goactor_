package actor

import (
	"fmt"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/sysmsg"
	"time"
)

type mailboxReceiver interface {
	Receive(handler func(message interface{}) (loop bool))
	ReceiveWithTimeout(d time.Duration, handler func(message interface{}) (loop bool))
}

type futureActor struct {
	pid ClosablePID
	receiver mailboxReceiver
	done bool
}

func NewFutureActor() *futureActor {
	m := mailbox.NewFutureMailbox()
	return &futureActor{
		pid: pid.NewFuturePID(m),
		receiver: m,
	}
}

func (f *futureActor) Self() UserPID {
	return f.pid
}

func (f *futureActor) Receive() (response interface{}, err error) {
	if f.done {
		err = mailbox.ErrDisposed
		return
	}
	defer f.Dispose()
	f.receiver.Receive(func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case sysmsg.Exit:
			err = fmt.Errorf("target Actor terminated before sending a response")
		case error:
			err = msg
		default:
			response = msg
		}
		return false
	})

	return
}

func (f *futureActor) ReceiveWithTimeout(duration time.Duration) (response interface{}, err error) {
	if duration < 0 {
		return f.Receive()
	} else if f.done {
		err = mailbox.ErrDisposed
		return
	}
	defer f.Dispose()
	f.receiver.ReceiveWithTimeout(duration, func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case sysmsg.Exit:
			err = fmt.Errorf("target Actor terminated before sending a response")
		case error:
			err = msg
		case sysmsg.Timeout:
			err = fmt.Errorf("timeout")
		default:
			response = msg
		}
		return false
	})

	return
}

func (f *futureActor) Dispose() {
	f.pid.Dispose()
	f.done = true
}