package sysmsg

type SystemMessage interface {
	systemMessage()
}

type Reason string

const (
	// Kill reason in case of a Shutdown message
	Kill   Reason = "kill"
	Panic  Reason = "panic"
	Normal Reason = "normal"
	SupMaxRestart Reason = "max_restarts_reached"
)

type Relation string

const (
	Linked    Relation = "linked"
	Monitored Relation = "monitored"
)
