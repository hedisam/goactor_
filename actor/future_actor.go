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

func (f *futureActor) Self() *pid.ProtectedPID {
	return pid.NewProtectedPID(f.pid)
}

func (f *futureActor) monitor(_pid *pid.ProtectedPID) {
	request := sysmsg.Monitor{Parent: pid.ExtractPID(_pid)}
	sendSystemMessage(_pid, request)
}

func (f *futureActor) Send(pid *pid.ProtectedPID, message interface{}) {
	f.monitor(pid)
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