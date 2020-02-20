package goactor

import "time"

type Mailbox interface {
	sendUserMessage(message interface{})
	sendSysMessage(message SystemMessage)
	close()
	receive(handler MessageHandler)
	receiveWithTimeout(duration time.Duration, handler MessageHandler)
	setActor(actor *Actor)
	getActor() *Actor
}