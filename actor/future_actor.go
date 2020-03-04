package actor

import (
	"fmt"
	"github.com/hedisam/goactor/internal/mailbox"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/sysmsg"
	"time"
)

type futureActor struct {
	pid pid.PID
}

func NewFutureActor() *futureActor {
	return &futureActor{
		pid: pid.NewFuturePID(),
	}
}

func (f *futureActor) Self() *PID {
	return &PID{f.pid}
}

func (f *futureActor) Monitor(pid *PID) {
	request := sysmsg.Monitor{Parent: f.Self().pid}
	sendSystemMessage(pid, request)
}

func (f *futureActor) SendAndMonitor(pid *PID, message interface{}) {
	f.Monitor(pid)
	Send(pid, message)
}

func (f *futureActor) Recv() (response interface{}, err error) {
	f.pid.Mailbox().Receive(func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case sysmsg.Exit:
			err = fmt.Errorf("target actor terminated before sending a response")
		case mailbox.ErrDisposed:
			err = fmt.Errorf("%v", msg)
		default:
			response = msg
		}
		return false
	})

	return
}

func (f *futureActor) RecvWithTimeout(duration time.Duration) (response interface{}, err error) {
	f.pid.Mailbox().ReceiveWithTimeout(duration, func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case sysmsg.Exit:
			err = fmt.Errorf("target actor terminated before sending a response")
		case mailbox.ErrDisposed:
			err = fmt.Errorf("%v", msg)
		case sysmsg.Timeout:
			err = fmt.Errorf("timeout")
		default:
			response = msg
		}
		return false
	})

	return
}