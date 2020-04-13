package spec


type BasicPID interface {
	ID() string
	SendUserMessage(message interface{})
	SendSystemMessage(message interface{})
}

type CancelablePID interface {
	BasicPID
	Shutdown()
}
