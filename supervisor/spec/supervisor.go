package spec

import (
	"fmt"
	"github.com/rs/xid"
)

type StartLink func(specs ...Spec) (*SupRef, error)

type SupervisorSpec struct {
	Id        string
	Children  []Spec
	StartLink StartLink
	Restart   int32
	Shutdown  int32
}

func NewSupervisorSpec(start StartLink, childSpecs ...Spec) SupervisorSpec {
	return SupervisorSpec{
		Id:        xid.New().String(),
		Children:  childSpecs,
		StartLink: start,
		Restart:   RestartTransient,
		Shutdown:  ShutdownKill,
	}
}

func (sup SupervisorSpec) Start(supervisor linkSpawner) (CancelablePID, error) {

}

func (sup SupervisorSpec) ChildSpec() SupervisorSpec {
	return sup
}

func (sup SupervisorSpec) SetId(id string) SupervisorSpec {
	sup.Id = id
	return sup
}

func (sup SupervisorSpec) SetRestart(restart int32) SupervisorSpec {
	sup.Restart = restart
	return sup
}

func (sup SupervisorSpec) SetShutdown(shutdown int32) SupervisorSpec {
	sup.Shutdown = shutdown
	return sup
}

func (sup SupervisorSpec) Type() int32 {
	return SupervisorActor
}

func (sup SupervisorSpec) ID() string {
	return sup.Id
}

func (sup SupervisorSpec) RestartValue() int32 {
	return sup.Restart
}

func (sup SupervisorSpec) Validate() error {
	if sup.Id == "" {
		return fmt.Errorf("childspec's id could not be empty")
	} else if sup.Restart != RestartAlways && sup.Restart != RestartTransient && sup.Restart != RestartNever {
		return fmt.Errorf("invalid childspec's restart value: %v, id %s", sup.Restart, sup.Id)
	} else if sup.Shutdown < ShutdownInfinity {
		return fmt.Errorf("invalid childspec's shutdown value: %v, id %s", sup.Shutdown, sup.Id)
	} else if sup.StartLink == nil {
		return fmt.Errorf("supervisor childspec's StartLink could not be nil, id %s", sup.Id)
	} else if sup.Children == nil || len(sup.Children) == 0 {
		return fmt.Errorf("supervisor child list is nil or empty, id: %s", sup.Id)
	}
	return nil
}
