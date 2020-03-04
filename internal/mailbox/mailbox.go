package mailbox

import (
	"time"
)

const (
	defaultUserMailboxCap = 100
	defaultSysMailboxCap  = 10
)

const (
	mailboxProcessing int32 = iota
	mailboxIdle
)

type Mailbox interface {
	SendUserMessage(message interface{})
	SendSystemMessage(message interface{})
	Receive(handler MessageHandler)
	ReceiveWithTimeout(d time.Duration, handler MessageHandler)
	Dispose()
	Utils() *ActorUtils
}

type ActorUtils struct {
	MonitoredBy func(pid interface{})
	DemonitorBy func(pid interface{})
	Link        func(pid interface{})
	Unlink      func(pid interface{})
	Self        func() (pid interface{})
	TrapExit    func() bool
}

type MessageHandler func(message interface{}) (loop bool)
