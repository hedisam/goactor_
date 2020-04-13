package actor

import (
	"context"
)

type SupervisorContext interface {
	Args() []interface{}
	Receive(handler func(message interface{}) (loop bool))
	Done() <-chan struct{}
	Context() context.Context
}

type supervisorActor struct {
	SupervisorContext
	trapExit int32
	self ClosablePID
	linkedActors *linkedActors
	mySupervisor UserPID
}

func NewSupervisorActor(ctx SupervisorContext) *supervisorActor {
	return &supervisorActor{
		SupervisorContext: ctx,
		trapExit:          trapExitYes,
		self:              nil,
		linkedActors:      &linkedActors{repo:connectedActorsRepository{}},
		mySupervisor:      nil,
	}
}

func (s *supervisorActor) trapExited() bool {
	return true
}

