package goactor

import (
	"github.com/Workiva/go-datastructures/queue"
	"log"
	"sync/atomic"
	"time"
)

const (
	actorPaused		int32 = iota
	actorActive
)

var pauseDuration = time.Second * 2
type PauseCMD struct {}

type queueMailboxPausable struct {
	userMailbox *queue.RingBuffer
	sysMailbox  *queue.RingBuffer
	done        chan struct{}
	actor       *Actor
	status      int32
	signal      chan struct{}
	workStatus	int32
}

// defaultAPRingBufferQueueMailbox returns an auto pausing ring buffer mailbox
func defaultAPRingBufferQueueMailbox() Mailbox {
	m := queueMailboxPausable{
		userMailbox:	queue.NewRingBuffer(defaultUserMailboxCap),
		sysMailbox:  	queue.NewRingBuffer(defaultSysMailboxCap),
		done:       	make(chan struct{}),
		status:      	mailboxIdle,
		signal:      	make(chan struct{}),
		workStatus:		actorActive,
	}
	return &m
}

func (m *queueMailboxPausable) setActor(actor *Actor) {
	m.actor = actor
}

func (m *queueMailboxPausable) getActor() *Actor {
	return m.actor
}

func (m *queueMailboxPausable) sendUserMessage(message interface{}) {
	select {
	case <-m.done:
		return
	default:
		//if atomic.CompareAndSwapInt32(&m.workStatus, actorPaused, actorActive) {
		//	// resume actor
		//	spawn(m.actor)
		//	log.Println("[+] actor resumed, id:", m.actor.pid.id)
		//}

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

func (m *queueMailboxPausable) sendSysMessage(message SystemMessage) {
	m.sendUserMessage(message)
	return
}

func (m *queueMailboxPausable) receive(handler MessageHandler) {
	//actorTimer := time.NewTimer(pauseDuration)
	//defer stopTimer(actorTimer)
listen:
	select {
	case <-m.done:
		return
	//case <-actorTimer.C:
	//	// sleep/pause
	//	atomic.StoreInt32(&m.workStatus, actorPaused)
	//	//m.fade()
	//	panic(PauseCMD{})
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
		//resetTimer(actorTimer, pauseDuration, false)
		goto listen
	}
}

func (m *queueMailboxPausable) receiveWithTimeout(d time.Duration, handler MessageHandler) {
	timer := time.NewTimer(d)
	defer stopTimer(timer)
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

func (m *queueMailboxPausable) close() {
	close(m.done)
}

// handleSystemMessage return true if the message should be passed to the user
func (m *queueMailboxPausable) handleSystemMessage(message interface{}) (bool, SystemMessage) {
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
		if trapExit == actorTrapExitYes {
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
