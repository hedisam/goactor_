package pid

import "github.com/hedisam/goactor/internal/mailbox"

type localPID struct {
	m mailbox.Mailbox
	shutdown func()
	actorType func(int32)
	supervisorSetter func(PID)
}

func NewPID(utils *mailbox.ActorUtils) *localPID {
	return &localPID{
		m: mailbox.DefaultRingBufferQueueMailbox(utils),
	}
}

func (pid *localPID) Mailbox() mailbox.Mailbox {
	return pid.m
}

func (pid *localPID) ShutdownFn() func() {
	return pid.shutdown
}

func (pid *localPID) SetShutdownFn(shutdown func()) {
	pid.shutdown = shutdown
}

func (pid *localPID) SetActorTypeFn(fn func(int32)) {
	pid.actorType = fn
}

func (pid *localPID) ActorTypeFn() func(int32) {
	return pid.actorType
}

func (pid *localPID) SetSupervisorFn(fn func(pid PID)) {
	pid.supervisorSetter = fn
}

func (pid *localPID) SupervisorFn() func(pid PID) {
	return pid.supervisorSetter
}
