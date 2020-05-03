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

type SupervisorActor struct {
	SupervisorContext
	trapExit int32
	self ClosablePID
	linkedActors *linkedActors
	mySupervisor UserPID
}

func NewSupervisorActor(ctx SupervisorContext, pid ClosablePID) *SupervisorActor {
	return &SupervisorActor{
		SupervisorContext: ctx,
		trapExit:          trapExitYes,
		self:              pid,
		linkedActors:      &linkedActors{repo:connectedActorsRepository{}},
		mySupervisor:      nil,
	}
}

func (s *SupervisorActor) SpawnMe(fn func(sup *SupervisorActor)) {
	// spawn
	go func() {
		defer s.handleTermination()
		fn(s)
	}()

}

func (s *SupervisorActor) Self() UserPID {
	return s.self
}

func (s *SupervisorActor) SpawnLink(fn Func, args ...interface{}) CancelablePID {
	actor := createActor(args...)
	actor.connectedActors.link(s.self)
	spawn(fn, actor)

	s.linkedActors.link(actor.self)
	return pid
}

func (s *SupervisorActor) UnlinkDeadChild(pid UserPID) {
	s.linkedActors.unlink(pid)
}

func (s *SupervisorActor) trapExited() bool {
	return true
}

func (s *SupervisorActor) handleTermination() {

}

