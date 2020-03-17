package spec

import (
	"fmt"
)

type SpecsMap map[string]Spec

func (sm SpecsMap) Restart(name string) int32 {
	switch spec := sm[name].(type) {
	case WorkerSpec:
		return spec.Restart
	case SupervisorSpec:
		return spec.Restart
	default:
		panic("invalid childspec type in SpecMap Restart")
	}
}

func (sm SpecsMap) Shutdown(name string) int32 {
	switch spec := sm[name].(type) {
	case WorkerSpec:
		return spec.Shutdown
	case SupervisorSpec:
		return spec.Shutdown
	default:
		panic("invalid childspec type in SpecMap Restart")
	}
}

func (sm SpecsMap) WorkerStartSpec(name string) *WorkerStartSpec {
	switch spec := sm[name].(type) {
	case WorkerSpec:
		return &spec.Start
	default:
		return nil
	}
}

func (sm SpecsMap) SupervisorStartLink(name string) StartLink {
	switch spec := sm[name].(type) {
	case SupervisorSpec:
		return spec.StartLink
	default:
		return nil
	}
}

func (sm SpecsMap) SupervisorChildren(name string) []Spec {
	switch spec := sm[name].(type) {
	case SupervisorSpec:
		return spec.Children
	default:
		return nil
	}
}

func (sm SpecsMap) Type(name string) ChildType {
	switch spec := sm[name].(type) {
	case SupervisorSpec:
		return spec.Type()
	case WorkerSpec:
		return spec.Type()
	default:
		return -1
	}
}

func ToMap(specs ...Spec) (specsMap SpecsMap, err error) {
	if len(specs) == 0 {
		err = fmt.Errorf("empty childspec list")
		return
	}

	specsMap = make(SpecsMap)
	for _, s := range specs {
		if s == nil {
			err = fmt.Errorf("childspec could not be nil")
			return
		}
		switch spc := s.ChildSpec().(type) {
		case WorkerSpec:
			if spc.Id == "" {
				err = fmt.Errorf("childspec's id could not be empty")
				return
			} else if _, duplicate := specsMap[spc.Id]; duplicate {
				err = fmt.Errorf("duplicate childspec id %s", spc.Id)
				return
			} else if spc.Restart != RestartAlways && spc.Restart != RestartTransient && spc.Restart != RestartNever {
				err = fmt.Errorf("invalid childspec's restart value: %v, id %s", spc.Restart, spc.Id)
				return
			} else if spc.Shutdown < ShutdownInfinity {
				err = fmt.Errorf("invalid childspec's shutdown value: %v, id %s", spc.Shutdown, spc.Id)
				return
			} else if spc.Start.ActorFunc == nil {
				err = fmt.Errorf("childspec's fn (actor.Func(actor.Actor)) could not be nil, id %s", spc.Id)
				return
			}
			specsMap[spc.Id] = s.ChildSpec()
		case SupervisorSpec:
			if spc.Id == "" {
				err = fmt.Errorf("childspec's id could not be empty")
				return
			} else if _, duplicate := specsMap[spc.Id]; duplicate {
				err = fmt.Errorf("duplicate childspec id %s", spc.Id)
				return
			} else if spc.Restart != RestartAlways && spc.Restart != RestartTransient && spc.Restart != RestartNever {
				err = fmt.Errorf("invalid childspec's restart value: %v, id %s", spc.Restart, spc.Id)
				return
			} else if spc.Shutdown < ShutdownInfinity {
				err = fmt.Errorf("invalid childspec's shutdown value: %v, id %s", spc.Shutdown, spc.Id)
				return
			} else if spc.StartLink == nil {
				err = fmt.Errorf("supervisor childspec's StartLink could not be nil, id %s", spc.Id)
				return
			} else if spc.Children == nil || len(spc.Children) == 0 {
				err = fmt.Errorf("supervisor child list is nil or empty, id: %s", spc.Id)
				return
			}
			specsMap[spc.Id] = s.ChildSpec()
		default:
			err = fmt.Errorf("invalid childspec type: %T %v", s, s)
			return
		}
	}
	return
}
