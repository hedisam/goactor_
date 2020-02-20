package goactor

import (
	"github.com/Workiva/go-datastructures/queue"
	"log"
	"sync/atomic"
	"time"
)

type queueMailbox struct {
	userMailbox *queue.RingBuffer
	sysMailbox  *queue.RingBuffer
	done        chan struct{}
	actor       *Actor
	status      int32
	signal      chan struct{}
}


func defaultQueueMailbox() Mailbox {
	m := queueMailbox{
		userMailbox:	queue.NewRingBuffer(defaultUserMailboxCap),
		sysMailbox:  	queue.NewRingBuffer(defaultSysMailboxCap),
		done:       	make(chan struct{}),
		status:      	mailboxIdle,
		signal:      	make(chan struct{}),
	}
	return &m
}

func (m *queueMailbox) setActor(actor *Actor) {
	m.actor = actor
}

func (m *queueMailbox) getActor() *Actor {
	return m.actor
}

func (m *queueMailbox) sendUserMessage(message interface{}) {
	select {
	case <-m.done:
		return
	default:
		// todo: should we return the error? probably yes
		err := m.userMailbox.Put(message)
		if err != nil {
			log.Println("queue_mailbox put error:", err, m.actor)
			return
		}
		if atomic.CompareAndSwapInt32(&m.status, mailboxIdle, mailboxProcessing) {
			select {
			case m.signal<- struct{}{}:
			case <-m.done:
				return
			}
		}
	}
}

func (m *queueMailbox) sendSysMessage(message SystemMessage) {
	m.sendUserMessage(message)
	return
}

func (m *queueMailbox) receive(handler MessageHandler) {
	// todo: handle sys messages separately
listen:
	select {
	case <-m.done:
		return
	case <-m.signal:
		for m.userMailbox.Len() != 0 {
			msg, _ := m.userMailbox.Get()
			switch msg.(type) {
			case SystemMessage:
				pass, msg := m.handleSystemMessage(msg)
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

func (m *queueMailbox) receiveWithTimeout(d time.Duration, handler MessageHandler) {
	timer := time.NewTimer(d)
listen:
	select {
	case <-m.done:
		return
	case <-m.signal:
		for m.userMailbox.Len() != 0 {
			msg, _ := m.userMailbox.Get()
			switch msg.(type) {
			case SystemMessage:
				pass, msg := m.handleSystemMessage(msg)
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
		keepOn := handler(TimeoutMessage{})
		if !keepOn {
			return
		}
		resetTimer(timer, d, true)
		goto listen
	}
}

func (m *queueMailbox) close() {
	close(m.done)
}

// handleSystemMessage return true if the message should be passed to the user
func (m *queueMailbox) handleSystemMessage(message interface{}) (bool, SystemMessage) {
	switch msg := message.(type) {
	// a monitored/linked actor has terminated with a normal status
	case NormalExit:
		return true, msg
	// a monitored or linked(trap_exit must be true, see case ExitCMD) actor has terminated by a panic
	case PanicExit:
		return true, msg
	// a monitored actor has terminated	by receiving an ExitCMD because of another linked actor panic
	case KillExit:
		return true, msg
	// a linked actor has terminated by panic, now, making us to terminate too.
	case ExitCMD:
		trapExit := atomic.LoadInt32(&m.actor.trapExit)
		if trapExit == actor_trap_exit_yes {
			// we are trapping exit commands from linked actors. so just convert them to a PanicExit message
			// which is a notifying message
			return true, PanicExit{who: msg.becauseOf, reason: msg.reason}
		} else {
			// not trapping exit commands. we have to terminate.
			// we are using panic to delegate termination handling to the deferred function that recovers.
			// notice: the user should not defer a closure that calls to recover().
			panic(msg)
		}
	case MonitorRequest:
		// todo: maybe send an act message to the sender. future.
		if msg.demonitor {
			delete(m.actor.monitorActors, msg.by.pid.id)
		} else {
			m.actor.monitorActors[msg.by.pid.id] = msg.by
		}
		return false, nil
	case LinkRequest:
		if msg.unlink {
			delete(m.actor.linkedActors, msg.to.pid.id)
		} else {
			m.actor.linkedActors[msg.to.pid.id] = msg.to
		}
		return false, nil
	default:
		log.Println("mailbox: unknown sys message", msg)
		return false, nil
	}
}
