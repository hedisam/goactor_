package mailbox

import (
	"github.com/hedisam/goactor/sysmsg"
	"time"
)

type channelMailbox struct {
	userMailbox chan interface{}
	sysMailbox  chan interface{}
	done        chan struct{}
	utils       *ActorUtils
}

func DefaultChanMailbox(utils *ActorUtils) Mailbox {
	m := channelMailbox{
		userMailbox: make(chan interface{}, defaultUserMailboxCap),
		sysMailbox:  make(chan interface{}, defaultSysMailboxCap),
		done:        make(chan struct{}),
		utils:       utils,
	}
	return &m
}

func (m *channelMailbox) Utils() *ActorUtils {
	return m.utils
}

func (m *channelMailbox) SendUserMessage(message interface{}) {
	select {
	case <-m.done:
		return
	case m.userMailbox <- message:
	}
}

func (m *channelMailbox) SendSystemMessage(message interface{}) {
	select {
	case <-m.done:
		return
	case m.sysMailbox <- message:
	}
}

func (m *channelMailbox) Receive(handler MessageHandler) {
	defer checkContext(m)
loop:
	select {
	case msg, ok := <-m.userMailbox:
		if !ok {
			return
		}
		keepOn := handler(msg)
		if keepOn {
			goto loop
		}
	case sysMsg := <-m.sysMailbox:
		pass, msg := handleSystemMessage(m, sysMsg)
		if pass {
			keepOn := handler(msg)
			if keepOn {
				goto loop
			}
		} else {
			goto loop
		}
	case <-m.done:
		// we're not accepting any messages
		return
	}
}

func (m *channelMailbox) ReceiveWithTimeout(timeout time.Duration, handler MessageHandler) {
	defer checkContext(m)
	timer := time.NewTimer(timeout)
	defer stopTimer(timer)
loop:
	select {
	case msg, ok := <-m.userMailbox:
		if !ok {
			return
		}
		keepOn := handler(msg)
		if keepOn {
			resetTimer(timer, timeout, false)
			goto loop
		}
	case sysMsg := <-m.sysMailbox:
		handleSystemMessage(m, sysMsg)
		resetTimer(timer, timeout, false)
		goto loop
	case <-m.done:
		return
	case <-timer.C:
		keepOn := handler(sysmsg.Timeout{})
		if keepOn {
			resetTimer(timer, timeout, true)
			goto loop
		}
	}

}

func (m *channelMailbox) Dispose() {
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
