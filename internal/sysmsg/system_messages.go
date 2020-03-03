package sysmsg

type SystemMessage interface {
	systemMessage()
}

type Reason string

const (
	Kill   Reason = "kill"
	Panic  Reason = "panic"
	Normal Reason = "normal"
)

type Relation string

const (
	Linked    Relation = "linked"
	Monitored Relation = "monitored"
)
