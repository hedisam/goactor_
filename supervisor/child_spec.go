package supervisor

import "github.com/hedisam/goactor/actor"

type ChildType int32
type childSpecMap map[string]ChildSpec

const (
	Worker ChildType = iota
	Supervisor
)

const (
	RestartAlways int32 = iota
	RestartTransient
	RestartNever
)

const (
	ShutdownInfinity int32 = iota - 1 // -1
	ShutdownKill                      // 0
	// >= 1 as number of milliseconds
)

type ChildSpec struct {
	id        string
	start     StartSpec
	restart   int32
	shutdown  int32
	childType ChildType
}

type StartSpec struct {
	actorFunc actor.Func
	args      []interface{}
}

func NewChildSpec(id string, fn actor.Func, args ...interface{}) *ChildSpec {
	return &ChildSpec{
		id:        id,
		start:     StartSpec{actorFunc: fn, args: args},
		restart:   RestartAlways,
		shutdown:  ShutdownKill,
		childType: Worker,
	}
}
