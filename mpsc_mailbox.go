package goactor

import (
	"github.com/t3rm1n4l/go-mpscqueue"
	"sync/atomic"
	"time"
)

const (
	mailboxProcessing int32 = iota
	mailboxIdle
)

type mpscMailbox struct {
	userMailbox *mpsc.MPSCQueue
	sysMailbox  *mpsc.MPSCQueue
	done        chan struct{}
	actor       *Actor
	status      int32
	signal      chan struct{}
}

func defaultMPSCMailbox() Mailbox {
	m := mpscMailbox{
		userMailbox: mpsc.New(),
		sysMailbox:  mpsc.New(),
		done:        make(chan struct{}),
		status:      mailboxIdle,
		signal:      make(chan struct{}),
	}
	return &m
}

func (m *mpscMailbox) setActor(actor *Actor) {
	m.actor = actor
}

func (m *mpscMailbox) getActor() *Actor {
	return m.actor
}

func (m *mpscMailbox) sendUserMessage(message interface{}) {
	select {
	case <-m.done:
		return
	default:
		m.userMailbox.Push(message)
		if atomic.CompareAndSwapInt32(&m.status, mailboxIdle, mailboxProcessing) {
			select {
			case m.signal<- struct{}{}:
			case <-m.done:
				return
			}
		}
	}
}

func (m *mpscMailbox) sendSysMessage(message SystemMessage) {
	// todo: complete the implementation
	select {
	case <-m.done:
		return
	default:
		m.sysMailbox.Push(message)
	}
}

func (m *mpscMailbox) receive(handler MessageHandler) {
	// todo: handle sys messages
	listen:
		select {
		case <-m.done:
			return
		case <-m.signal:
			for m.userMailbox.Size() != 0 {
				keepOn := handler(m.userMailbox.Pop())
				if !keepOn {
					return
				}
			}
			atomic.StoreInt32(&m.status, mailboxIdle)
			goto listen
		}
}

func (m *mpscMailbox) receiveWithTimeout(d time.Duration, handler MessageHandler) {
	return
}

func (m *mpscMailbox) close() {
	close(m.done)
}