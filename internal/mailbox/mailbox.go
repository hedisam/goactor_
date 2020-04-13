package mailbox

const (
	defaultUserMailboxCap = 100
	defaultSysMailboxCap  = 10
)

const (
	mailboxProcessing int32 = iota
	mailboxIdle
)

type systemMessageHandler interface {
	HandleSystemMessage(message interface{}) (passToUser bool, msg interface{})
	CheckUnhandledShutdown()
}

type MessageHandler func(message interface{}) (loop bool)
