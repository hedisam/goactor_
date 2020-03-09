package spec

import (
	"github.com/hedisam/goactor/actor"
)

type WorkerSpec struct {
	Id        string
	Start     WorkerStartSpec
	Restart   int32
	Shutdown  int32
}

type WorkerStartSpec struct {
	ActorFunc actor.Func
	Args      []interface{}
}

func (w WorkerSpec) ChildSpec() Spec {
	return w
}

func (w WorkerSpec) SetRestart(restart int32) WorkerSpec {
	w.Restart = restart
	return w
}

func (w WorkerSpec) SetShutdown(shutdown int32) WorkerSpec {
	w.Shutdown = shutdown
	return w
}
