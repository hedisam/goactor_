package actor

type UserPID interface {
	ID() string
	SendUserMessage(message interface{})
	SendSystemMessage(message interface{})
}

type ClosablePID interface {
	UserPID
	Dispose()
}

type CancelablePID interface {
	UserPID
	Shutdown()
}