package pid

import "github.com/hedisam/goactor/internal/mailbox"

type futurePID struct {
	mailbox mailbox.Mailbox
	shutdown func()
	actorType func(int32)
}

func NewFuturePID() PID {
	return &futurePID{
		mailbox: mailbox.NewFutureMailbox(),
	}
}

func (f *futurePID) Mailbox() mailbox.Mailbox {
	return f.mailbox
}

func (f *futurePID) ShutdownFn() func() {
	return f.shutdown
}

// we don't need the following methods for the future actor

func (f *futurePID) SetShutdownFn(shutdown func()) {
	f.shutdown = shutdown
}

func (f *futurePID) SetActorTypeFn(fn func(actorType int32)) {
	f.actorType = fn
}

func (f *futurePID) ActorTypeFn() func(int32) {
	return f.actorType
}

func (f *futurePID) SetSupervisorFn(fn func(pid PID)) {
}

func (f *futurePID) SupervisorFn() func(pid PID) {
	return nil
}
