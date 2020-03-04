package pid

import "github.com/hedisam/goactor/internal/mailbox"

type futurePID struct {
	mailbox mailbox.Mailbox
	shutdown func()
}

func NewFuturePID() PID {
	return &futurePID{
		mailbox: mailbox.NewFutureMailbox(),
	}
}

func (f *futurePID) Mailbox() mailbox.Mailbox {
	return f.mailbox
}

func (f *futurePID) Shutdown() func() {
	return f.shutdown
}

func (f *futurePID) SetShutdown(shutdown func()) {
	f.shutdown = shutdown
}
