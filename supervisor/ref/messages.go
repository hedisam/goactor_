package ref

import (
	"fmt"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/supervisor/spec"
)

type Error error

func errInvalidResponse(resp interface{}) error {
	return fmt.Errorf("supervisor has sent invalid response: %v", resp)
}

type Call struct {
	Sender *pid.ProtectedPID
	Request interface{}
}

// OK represents a successful result
type OK struct {

}

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
	Spec spec.Spec
}

type Stop struct {
	Reason string
}

type TerminateChild struct {
	Id string
}

type WithChildren struct {
	ChildrenInfo []spec.ChildInfo
}