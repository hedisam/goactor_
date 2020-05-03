package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/supervisor/spec"
)

type specsMap map[string]spec.ChildSpec

func toMap(specs ...spec.Spec) (specsMap, error) {
	if len(specs) == 0 {
		return nil, fmt.Errorf("empty childspec list")
	}

	specsMap := make(specsMap)
	for _, s := range specs {
		if s == nil {
			return nil, fmt.Errorf("childspec could not be nil")
		}

		err := s.ChildSpec().Validate()
		if err != nil {
			return nil, err
		}

		if _, duplicate := specsMap[s.ChildSpec().ID()]; duplicate {
			return nil, fmt.Errorf("duplicate childspec id %s", s.ChildSpec().ID())
		}

		specsMap[s.ChildSpec().ID()] = s.ChildSpec()
	}

	return specsMap, nil
}
