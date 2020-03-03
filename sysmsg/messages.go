package sysmsg

import (
	"time"
)

// Exit describes an exit event emitted by a monitored/linked actor
type Exit struct {
	// Who is the actor that terminated
	Who interface{}
	// Parent is the actor that made "Who" to terminate
	Parent interface{}
	// Reason behind the termination
	Reason Reason
	// Relation describes the relationship between terminated actor and the one who received the message
	Relation Relation
}

func (e Exit) systemMessage() {}

// Shutdown is command omitted by a supervisor to terminate a supervised actor
type Shutdown struct {
	// Parent is the commanding actor/supervisor
	Parent interface{}
	// see supervisor shutdown values
	Shutdown int32
}

func (s Shutdown) systemMessage() {}

// Monitor describes a request sent to an actor to be monitored/demonitor by the parent
type Monitor struct {
	Parent interface{}
	// Revert is true when we ask to get demonitor-ed from parent
	Revert bool
}

func (m Monitor) systemMessage() {}

// Link describes a request sent to an actor to get linked with another one
type Link struct {
	To interface{}
	// Revert is true when we ask to get unlinked
	Revert bool
}

func (l Link) systemMessage() {}

type Timeout struct {
	Duration time.Duration
}

func (t Timeout) systemMessage() {}
