package supervisor

import (
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/supervisor/spec"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type state struct {
	specs      spec.SpecsMap
	options    *Options
	registry   *registry
	supervisor actor.Actor
}

func newState(specs spec.SpecsMap, options *Options, supervisor actor.Actor) *state {
	return &state{
		specs: specs,
		options: options,
		registry: newRegistry(options),
		supervisor: supervisor,
	}
}

func (state *state) shutdown(name string, _pid pid.PID) {
	// todo: do not call deadAndUnlink if we're supposed to receive shutdown feedback
	state.deadAndUnlink(_pid)

	actor.Send(pid.NewProtectedPID(_pid), sysmsg.Shutdown{
		Parent:   pid.ExtractPID(state.supervisor.Self()),
		Shutdown: state.specs.Shutdown(name),
	})
	_pid.ShutdownFn()()
}

func (state *state) maxRestartsReached() {
	// shutdown all spec and also the supervisor.
	// note: calling panic in supervisor should kill its children since they are linked but we're explicitly
	// shutting down each one to close child's context's done channel
	reg := copyMap(state.registry.aliveActors)
	for _pid, id := range reg {
		state.shutdown(id, _pid)
	}

	log.Println("[!] supervisor reached max restarts")
	panic(sysmsg.Exit{
		Who:      pid.ExtractPID(state.supervisor.Self()),
		Parent:   nil,
		Reason:   sysmsg.SupMaxRestart,
		Relation: sysmsg.Linked,
	})
}

func (state *state) spawn(name string) (err error) {
	if state.registry.reachedMaxRestarts(name) {
		state.maxRestartsReached()
		return
	}

	var ppid *pid.ProtectedPID
	switch state.specs.Type(name) {
	case spec.TypeWorker:
		start := state.specs.WorkerStartSpec(name)
		ppid = state.supervisor.SpawnLink(start.ActorFunc, start.Args...)
	case spec.TypeSupervisor:
		startLink := state.specs.SupervisorStartLink(name)
		ref, err := startLink(state.specs.SupervisorChildren(name)...)
		if err != nil {return err}
		ppid = ref.PPID
		state.supervisor.Link(ppid)
	default:
		panic("invalid spec type when spawning child")
	}
	pid.ExtractPID(ppid).SupervisorFn()(pid.ExtractPID(state.supervisor.Self()))
	// register locally
	state.registry.put(pid.ExtractPID(ppid), name)
	// register globally
	actor.Register(name, ppid)
	return
}

func (state *state) init() (err error) {
	for id := range state.specs {
		err = state.spawn(id)
		if err != nil {
			return
		}
	}
	return
}


func (state *state) handleOneForAll(name string) {
	reg := copyMap(state.registry.aliveActors)
	for _pid, id := range reg {
		_pid := _pid
		if id == name {
			// this actor already has been terminated so no need to shut it down
			// but we need to unlink and declare it dead
			state.deadAndUnlink(_pid)
		} else {
			state.shutdown(id, _pid)
		}
		state.spawn(id)
	}
}

func (state *state) handleOneForOne(name string, _pid pid.PID) {
	// we need to unlink the terminated actor and declare it dead
	state.deadAndUnlink(_pid)
	// re-spawn
	state.spawn(name)
}

func (state *state) handleRestForOne(name string) {
	log.Println("supervisor: rest_for_one Strategy")
}

func (state *state) deadAndUnlink(_pid pid.PID) {
	state.registry.dead(_pid)
	state.supervisor.Unlink(pid.NewProtectedPID(_pid))
}

func copyMap(src map[pid.PID]string) (dst map[pid.PID]string) {
	dst = make(map[pid.PID]string)
	for k, v := range src {
		dst[k] = v
	}
	return
}