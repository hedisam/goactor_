package supervisor

import (
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type state struct {
	specs      childSpecMap
	options    *Options
	registry   *registry
	supervisor actor.Actor
}

func newState(specs childSpecMap, options *Options, supervisor actor.Actor) *state {
	return &state{
		specs: specs,
		options: options,
		registry: newRegistry(options),
		supervisor: supervisor,
	}
}

func (state *state) shutdown(name string, _pid pid.PID) {
	state.registry.dead(_pid)
	actor.Send(pid.NewProtectedPID(_pid), sysmsg.Shutdown{
		Parent:   pid.ExtractPID(state.supervisor.Self()),
		Shutdown: state.specs[name].Shutdown,
	})
	_pid.Shutdown()()
}

func (state *state) maxRestartsReached() {
	// shutdown all specs and also the supervisor. restart the supervisor only if we have a child
	// with restart value set to RestartAlways
	// note: calling panic in supervisor should kill its specs since they are linked but we're explicitly
	// shutting down each one to close specs's context's done channel
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

func (state *state) spawn(name string) {
	if state.registry.reachedMaxRestarts(name) {
		state.maxRestartsReached()
		return
	}

	childPID := state.supervisor.SpawnLink(state.specs[name].Start.ActorFunc, state.specs[name].Start.Args...)
	// register locally
	state.registry.put(pid.ExtractPID(childPID), name)
	// register globally
	actor.Register(name, childPID)
}

func (state *state) init() {
	for id := range state.specs {
		state.spawn(id)
	}
}


func (state *state) handleOneForAll(name string) {
	reg := copyMap(state.registry.aliveActors)
	for _pid, id := range reg {
		_pid := _pid
		if id == name {
			// this actor already been terminated so no need to shut it down but we need to declare it as dead
			state.registry.dead(_pid)
		} else {
			state.shutdown(id, _pid)
		}
		state.spawn(id)
	}
}

func (state *state) handleRestForOne(name string) {
	log.Println("supervisor: rest_for_one Strategy")
}

func copyMap(src map[pid.PID]string) (dst map[pid.PID]string) {
	dst = make(map[pid.PID]string)
	for k, v := range src {
		dst[k] = v
	}
	return
}