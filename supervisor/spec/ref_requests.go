package spec

import (
	"fmt"
)

func errInvalidResponse(resp interface{}) error {
	return fmt.Errorf("invalid response sent by supervisor: %v", resp)
}

type Call struct {
	Sender BasicPID
	Request interface{}
}

// OK represents a successful result
type OK struct {}

type CountChildren struct {
	// the total count of children, dead or alive
	Specs int
	// the count of all actively running child processes managed by this supervisor
	Active int
	// the count of all children marked as child_type = supervisor in the specification list,
	// regardless if the child process is still alive
	Supervisors int
	// the count of all children marked as child_type = worker in the specification list,
	// regardless if the child process is still alive
	Workers int
}

type DeleteChild struct {
	Id string
}

type RestartChild struct {
	Id string
}

type StartChild struct {
	Spec Spec
}

type Stop struct {
	Reason string
}

type TerminateChild struct {
	Id string
}

type WithChildren struct {
	ChildrenInfo []ChildInfo
}