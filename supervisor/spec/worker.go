package spec

import (
	"fmt"
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

func (w WorkerSpec) ChildSpec() WorkerSpec {
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

func (w WorkerSpec) Type() int32 {
	return WorkerActor
}

func (w WorkerSpec) RestartValue() int32 {
	return w.Restart
}

func (w WorkerSpec) ID() string {
	return w.Id
}

func (w WorkerSpec) Validate() error {
	if w.Id == "" {
		return fmt.Errorf("childspec's id could not be empty")
	} else if w.Restart != RestartAlways && w.Restart != RestartTransient && w.Restart != RestartNever {
		return fmt.Errorf("invalid childspec's restart value: %v, id %s", w.Restart, w.Id)
	} else if w.Shutdown < ShutdownInfinity {
		return fmt.Errorf("invalid childspec's shutdown value: %v, id %s", w.Shutdown, w.Id)
	} else if w.Start.ActorFunc == nil {
		return fmt.Errorf("childspec's fn (actor.Func(actor.Actor)) could not be nil, id %s", w.Id)
	}
	return nil
}
