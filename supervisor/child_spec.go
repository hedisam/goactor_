package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
)

type ChildType int32
type childSpecMap map[string]ChildSpec

const (
	TypeWorker ChildType = iota
	TypeSupervisor
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
	Id        string
	Start     StartSpec
	Restart   int32
	Shutdown  int32
	ChildType ChildType
}

type StartSpec struct {
	ActorFunc actor.Func
	Args      []interface{}
}

func NewChildSpec(id string, fn actor.Func, args ...interface{}) ChildSpec {
	return ChildSpec{
		Id:        id,
		Start:     StartSpec{ActorFunc: fn, Args: args},
		Restart:   RestartTransient,
		Shutdown:  ShutdownKill,
		ChildType: TypeWorker,
	}
}

func (spec ChildSpec) SetRestart(restart int32) ChildSpec {
	spec.Restart = restart
	return spec
}

func (spec ChildSpec) SetShutdown(shutdown int32) ChildSpec {
	spec.Shutdown = shutdown
	return spec
}

func (spec ChildSpec) SetChildType(t ChildType) ChildSpec{
	spec.ChildType = t
	return spec
}

func specsToMap(specs []ChildSpec) (specsMap childSpecMap, err error) {
	if len(specs) == 0 {
		err = fmt.Errorf("empty childspec list")
		return
	}
	specsMap = make(childSpecMap)
	for _, s := range specs {
		if s.Id == "" {
			err = fmt.Errorf("childspec's id could not be empty")
			return
		} else if s.Start.ActorFunc == nil {
			err = fmt.Errorf("childspec's fn (actor.Func(actor.Actor)) could not be nil, id %s", s.Id)
			return
		} else if _, duplicate := specsMap[s.Id]; duplicate {
			err = fmt.Errorf("duplicate childspec id %s", s.Id)
			return
		} else if s.Restart != RestartAlways && s.Restart != RestartTransient && s.Restart != RestartNever {
			err = fmt.Errorf("invalid childspec's restart value: %v, id %s", s.Restart, s.Id)
			return
		} else if s.Shutdown < ShutdownInfinity {
			err = fmt.Errorf("invalid childspec's shutdown value: %v, id %s", s.Shutdown, s.Id)
			return
		} else if s.ChildType != TypeWorker && s.ChildType != TypeSupervisor {
			err = fmt.Errorf("invalid child type: %v, id %s", s.ChildType, s.Id)
			return
		}

		specsMap[s.Id] = s
	}
	return
}