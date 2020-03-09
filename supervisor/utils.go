package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/supervisor/spec"
)

func specsToMap1(specs []spec.Spec) (specsMap spec.SpecsMap, err error) {
	if len(specs) == 0 {
		err = fmt.Errorf("empty childspec list")
		return
	}

	specsMap = make(spec.SpecsMap)
	for _, s := range specs {
		switch spc := s.(type) {
		case spec.WorkerSpec:
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
			specsMap[spc.Id] = s
		case spec.SupervisorSpec:
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
			}
			specsMap[spc.Id] = s
		default:
			err = fmt.Errorf("invalid childspec type: %T %v", s, s)
			return
		}
	}
	return
}


