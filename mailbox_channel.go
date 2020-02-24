package goactor

import (
	"log"
	"runtime"
	"sync/atomic"
	"time"
)

const (
	defaultUserMailboxCap	= 100
	defaultSysMailboxCap	= 10
)

type channelMailbox struct {
	userMailbox chan interface{}
	sysMailbox  chan SystemMessage
	done        chan struct{}
	actor       *Actor
}

func defaultChanMailbox() Mailbox {
	m := channelMailbox{
		userMailbox: make(chan interface{}, defaultUserMailboxCap),
		sysMailbox:  make(chan SystemMessage, defaultSysMailboxCap),
		done:      make(chan struct{}),
	}
	return &m
}

func (m *channelMailbox) setActor(actor *Actor) {
	m.actor = actor
}

func (m *channelMailbox) getActor() *Actor {
	return m.actor
}

func (m *channelMailbox) sendUserMessage(message interface{}) {
	select {
	case <-m.done:
		return
	case m.userMailbox <- message:
	}
}

func (m *channelMailbox) sendSysMessage(message SystemMessage) {
	select {
	case <-m.done:
		return
	case m.sysMailbox<- message:
	}
}

func (m *channelMailbox) receive(handler MessageHandler) {
	loop:
		select {
		case msg, ok := <- m.userMailbox:
			if !ok {return}
			elapsedMilli, keepOn := measure(handler, msg)
			if keepOn {
				// todo: implement a better algorithm to run runtime.Gosched()
				if elapsedMilli > 999 {
					runtime.Gosched()
				}
				goto loop
			}
		case sysMsg := <-m.sysMailbox:
			m.handleSystemMessage(sysMsg)
			goto loop
		case <-m.done:
			// we're not accepting any messages
			return
		}
}

func (m *channelMailbox) handleSystemMessage(message SystemMessage) {
	// todo: invoke user handler directly if needed instead of sending message to userMailbox
	switch msg := message.(type) {
	// a monitored/linked actor has terminated with a normal status
	case NormalExit:
		m.userMailbox<- msg
	// a monitored or linked(trap_exit must be true, see case ExitCMD) actor has terminated by a panic
	case PanicExit:
		m.userMailbox<- msg
	// a monitored actor has terminated	by receiving an ExitCMD because of another linked actor panic
	case KillExit:
		m.userMailbox<- msg
	// a linked actor has terminated by panic, now, making us to terminate too.
	case ExitCMD:
		trapExit := atomic.LoadInt32(&m.actor.trapExit)
		if trapExit == actorTrapExitYes {
			// we are trapping exit commands from linked actors. so just convert them to a PanicExit message
			// which is a notifying message
			m.userMailbox<- PanicExit{who: msg.becauseOf, reason: msg.reason}
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
	case LinkRequest:
		if msg.unlink {
			delete(m.actor.linkedActors, msg.to.pid.id)
		} else {
			m.actor.linkedActors[msg.to.pid.id] = msg.to
		}
	default:
		log.Println("mailbox: unknown sys message", msg)
	}
}

func (m *channelMailbox) receiveWithTimeout(timeout time.Duration, handler MessageHandler) {
	timer := time.NewTimer(timeout)
	defer stopTimer(timer)
	loop:
		select {
		case msg, ok := <- m.userMailbox:
			if !ok {return}
			elapsedMilli, keepOn := measure(handler, msg)
			if keepOn {
				// todo: implement a better algorithm to run runtime.Gosched()
				// todo: maybe check the count of received messages without been blocked by the loop
				if elapsedMilli > 999 {
					runtime.Gosched()
				}
				resetTimer(timer, timeout, false)
				goto loop
			}
		case sysMsg := <-m.sysMailbox:
			m.handleSystemMessage(sysMsg)
			resetTimer(timer, timeout, false)
			goto loop
		case <-m.done:
			return
		case <-timer.C:
			elapsedMilli, keepOn := measure(handler, TimeoutMessage{})
			if keepOn {
				if elapsedMilli > 999 {
					runtime.Gosched()
				}
				resetTimer(timer, timeout, true)
				goto loop
			}
		}

}

func (m *channelMailbox) close() {
	close(m.done)
}

func resetTimer(timer *time.Timer, d time.Duration, triggered bool) {
	if !triggered {
		stopTimer(timer)
	}
	timer.Reset(d)
}

// deprecated. it's blocking
func stopTimer(timer *time.Timer) {
	// drain the channel
	if !timer.Stop() {
		<-timer.C
	}
}

func measure(fn MessageHandler, arg interface{}) (elapsedMilli int64, fnResult bool) {
	now := time.Now()
	fnResult = fn(arg)
	elapsedMilli = time.Since(now).Milliseconds()
	return
}