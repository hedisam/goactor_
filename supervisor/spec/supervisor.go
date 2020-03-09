package spec

type StartLink func() (*Ref, error)

type SupervisorSpec struct {
	Id        string
	Children  []Spec
	StartLink StartLink
	Restart   int32
	Shutdown  int32
}

func (w SupervisorSpec) ChildSpec() Spec {
	return w
}

func (w SupervisorSpec) SetRestart(restart int32) SupervisorSpec {
	w.Restart = restart
	return w
}

func (w SupervisorSpec) SetShutdown(shutdown int32) SupervisorSpec {
	w.Shutdown = shutdown
	return w
}
