package spec

import (
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

func (sup SupervisorSpec) ChildSpec() Spec {
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

func (sup SupervisorSpec) Type() ChildType {
	return TypeSupervisor
}
