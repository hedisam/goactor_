package pid

import "github.com/hedisam/goactor/internal/mailbox"

type futurePID struct {
	mailbox mailbox.Mailbox
}

func NewFuturePID() PID {
	return &futurePID{
		mailbox: mailbox.NewFutureMailbox(),
	}
}

func (f *futurePID) Mailbox() mailbox.Mailbox {
	return f.mailbox
}
