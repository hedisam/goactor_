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

func NewWorkerSpec(name string, fn actor.Func, args ...interface{}) WorkerSpec {
	return WorkerSpec{
		Id:       name,
		Start:    WorkerStartSpec{ActorFunc: fn, Args: args},
		Restart:  0,
		Shutdown: 0,
	}
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

func (w WorkerSpec) Type() ChildType {
	return WorkerActor
}
