package mailbox

import (
	"github.com/Workiva/go-datastructures/queue"
	"github.com/hedisam/goactor/sysmsg"
	"log"
	"sync/atomic"
	"time"
)

type queueMailbox struct {
	userMailbox *queue.RingBuffer
	sysMailbox  *queue.RingBuffer
	done        chan struct{}
	status      int32
	signal      chan struct{}
	utils       *ActorUtils
}

func DefaultRingBufferQueueMailbox(utils *ActorUtils) Mailbox {
	m := queueMailbox{
		userMailbox: queue.NewRingBuffer(defaultUserMailboxCap),
		sysMailbox:  queue.NewRingBuffer(defaultSysMailboxCap),
		done:        make(chan struct{}),
		status:      mailboxIdle,
		signal:      make(chan struct{}),
		utils:       utils,
	}
	return &m
}

func (m *queueMailbox) Utils() *ActorUtils {
	return m.utils
}

func (m *queueMailbox) SendUserMessage(message interface{}) {
	select {
	case <-m.done:
		return
	default:
		// todo: should we return the error? probably yes
		err := m.userMailbox.Put(message)
		if err != nil {
			log.Println("queue_mailbox put error:", err)
			return
		}
		if atomic.CompareAndSwapInt32(&m.status, mailboxIdle, mailboxProcessing) {
			select {
			case m.signal <- struct{}{}:
			case <-m.done:
				return
			}
		}
	}
}

func (m *queueMailbox) SendSystemMessage(message interface{}) {
	m.SendUserMessage(message)
}

func (m *queueMailbox) Receive(handler MessageHandler) {
	// todo: handle sys messages separately
listen:
	select {
	case <-m.done:
		return
	case <-m.signal:
		for m.userMailbox.Len() != 0 {
			msg, _ := m.userMailbox.Get()
			switch msg.(type) {
			case sysmsg.SystemMessage:
				pass, msg := handleSystemMessage(m, msg)
				if pass {
					keepOn := handler(msg)
					if !keepOn {
						atomic.StoreInt32(&m.status, mailboxIdle)
						return
					}
				}
			default:
				keepOn := handler(msg)
				if !keepOn {
					atomic.StoreInt32(&m.status, mailboxIdle)
					return
				}
			}
		}
		atomic.StoreInt32(&m.status, mailboxIdle)
		goto listen
	}
}

func (m *queueMailbox) ReceiveWithTimeout(d time.Duration, handler MessageHandler) {
	timer := time.NewTimer(d)
listen:
	select {
	case <-m.done:
		return
	case <-m.signal:
		for m.userMailbox.Len() != 0 {
			msg, _ := m.userMailbox.Get()
			switch msg.(type) {
			case sysmsg.SystemMessage:
				pass, msg := handleSystemMessage(m, msg)
				if pass {
					keepOn := handler(msg)
					if !keepOn {
						atomic.StoreInt32(&m.status, mailboxIdle)
						return
					}
				}
			default:
				keepOn := handler(msg)
				if !keepOn {
					atomic.StoreInt32(&m.status, mailboxIdle)
					return
				}
			}
		}
		atomic.StoreInt32(&m.status, mailboxIdle)
		resetTimer(timer, d, false)
		goto listen
	case <-timer.C:
		keepOn := handler(sysmsg.Timeout{})
		if !keepOn {
			return
		}
		resetTimer(timer, d, true)
		goto listen
	}
}

func (m *queueMailbox) Dispose() {
	close(m.done)
}
