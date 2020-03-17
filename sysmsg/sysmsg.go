package sysmsg

type SystemMessage interface {
	systemMessage()
}

type Reason struct {
	Type string
	Details interface{}
}

const (
	// Kill reason in case of a Shutdown message
	Kill          = "kill"
	Panic         = "panic"
	Normal        = "normal"
	SupMaxRestart = "sup_reached_max_restarts"
)

type Relation string

const (
	Linked    Relation = "linked"
	Monitored Relation = "monitored"
)
