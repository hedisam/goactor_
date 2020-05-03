package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/supervisor/spec"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type state struct {
	specs      specsMap
	options    *Options
	registry   *registry
	supervisor *actor.SupervisorActor
}

func newState(specs specsMap, options *Options, supervisor *actor.SupervisorActor) *state {
	return &state{
		specs: specs,
		options: options,
		registry: newRegistry(options),
		supervisor: supervisor,
	}
}

func (state *state) shutdown(pid spec.CancelablePID) {
	// note: do not call deadAndUnlink if we're supposed to receive shutdown feedback
	state.deadAndUnlink(pid)

	actor.Send(pid, sysmsg.Shutdown{
		Parent:   state.supervisor.Self(),
		Shutdown: 0,
	})
	pid.Shutdown()
}

func (state *state) shutdownSupervisor(reason sysmsg.Reason) {
	// shutdown all specs then panic.
	// note: calling panic in supervisor should kill its children since they are linked but we're explicitly
	// shutting down each one to close child's context's done channel
	reg := copyMap(state.registry.aliveActors)
	for pid := range reg {
		state.shutdown(pid)
	}

	panic(sysmsg.Exit{
		Who:      state.supervisor.Self(),
		Parent:   nil,
		Reason:   reason,
	})
}

func (state *state) spawn(name string) error {
	if state.registry.reachedMaxRestarts(name) {
		log.Println("[!] supervisor reached max restarts")
		state.shutdownSupervisor(
			sysmsg.Reason{
				Type:    sysmsg.SupMaxRestart,
				Details: "one of supervisor's children reached its max allowed restarts",
			},
		)
	}

	var pid spec.CancelablePID
	s := state.specs[name]
	switch s := state.specs[name].(type) {
	case spec.WorkerSpec:
		start := s.Start
		pid = state.supervisor.SpawnLink(start.ActorFunc, start.Args...)
	case spec.SupervisorSpec:
		startLink := s.StartLink
		supRef, err := startLink(state.specs.SupervisorChildren(name)...)
		if err != nil {return err}
		pid = supRef.PID
		state.supervisor.Link(pid)
	default:
		log.Fatal("invalid spec type when spawning child")
	}

	// tell the new spawned actor (worker/supervisor) that it has a supervisor
	setSupervisor := pid.SupervisorFn()
	setSupervisor(state.supervisor.Self())

	// register locally
	state.registry.put(pid, name)
	// register globally
	process.Register(name, pid)
	return nil
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
			state.shutdown(_pid)
		}
		_ = state.spawn(id)
	}
}

func (state *state) handleOneForOne(name string, pid spec.BasicPID) {
	// we need to unlink the terminated actor and declare it dead
	state.deadAndUnlink(pid)
	// re-spawn
	_ = state.spawn(name)
}

func (state *state) handleRestForOne(_ string) {
	log.Println("supervisor: rest_for_one Strategy")
}

func (state *state) deadAndUnlink(pid spec.BasicPID) {
	state.registry.dead(pid)
	state.supervisor.UnlinkDeadChild(pid)
}

func (state *state) handleCall(call spec.Call) bool {
	switch request := call.Request.(type) {
	case spec.CountChildren:
		request.Specs = len(state.specs)
		request.Active = len(state.registry.aliveActors)
		for id := range state.specs {
			if state.specs.Type(id) == spec.SupervisorActor {
				request.Supervisors++
			} else {
				request.Workers++
			}
		}
		actor.Send(call.Sender, request)
	case spec.DeleteChild:
		// check if a child exists with the specified id
		if _, exists := state.specs[request.Id]; !exists {
			actor.Send(call.Sender, fmt.Errorf("child does not exists"))
			return true
		}
		// check if the child is running
		_, alive := state.registry.alivePID(request.Id)
		if alive {
			// we can not delete a child that is running
			actor.Send(call.Sender, fmt.Errorf("running child cannot be deleted"))
			return true
		}
		// delete the child
		// note: if this supervisor gets restarted by a parent supervisor then the original child specs
		// will be used. (unless we update the parent with the new child specs)
		delete(state.specs, request.Id)
		actor.Send(call.Sender, spec.OK{})
	case spec.RestartChild:
		// check if a child exists with the specified id
		if _, exists := state.specs[request.Id]; !exists {
			actor.Send(call.Sender, fmt.Errorf("child does not exists"))
			return true
		}
		// check if the child is running
		_, alive := state.registry.alivePID(request.Id)
		if alive {
			// we can not delete a child that is running
			actor.Send(call.Sender, fmt.Errorf("running child cannot be deleted"))
			return true
		}

		err := state.spawn(request.Id)
		if err != nil {
			actor.Send(call.Sender, err)
			return true
		}
		actor.Send(call.Sender, spec.OK{})
	case spec.StartChild:
		// check if the child spec is valid
		specMap, err := spec.ToMap(request.Spec)
		if err != nil {
			actor.Send(call.Sender, err)
			return true
		}
		// check if we've already got a child spec with the same id
		var id string
		for id = range specMap {}
		if _, exists := state.specs[id]; exists {
			actor.Send(call.Sender, fmt.Errorf("a child spec already present with the same id"))
			return true
		}
		// add the child spec to the supervisor child spec map
		state.specs[id] = specMap[id]
		// start the child
		err = state.spawn(id)
		if err != nil {
			actor.Send(call.Sender, err)
			return true
		}
		actor.Send(call.Sender, spec.OK{})
	case spec.Stop:
		// todo: pass the reason, make sure it's valid
		// shutdown children
		reg := copyMap(state.registry.aliveActors)
		for _pid, id := range reg {
			state.shutdown(_pid)
		}
		actor.Send(call.Sender, spec.OK{})
		return false
	case spec.TerminateChild:
		// check if a child exists with the specified id
		if _, exists := state.specs[request.Id]; !exists {
			actor.Send(call.Sender, fmt.Errorf("child does not exists"))
			return true
		}
		// check if the child is running
		_pid, ok := state.registry.alivePID(request.Id)
		if !ok {
			// the child is not alive
			actor.Send(call.Sender, fmt.Errorf("child already has been terminated"))
			return true
		}
		state.shutdown(_pid)
		actor.Send(call.Sender, spec.OK{})
	case spec.WithChildren:
		getPID := func(id string) spec.CancelablePID {
			pid, ok := state.registry.alivePID(id)
			if !ok {
				return nil
			}
			return pid
		}
		info := make([]spec.ChildInfo, 0, len(state.specs))
		for id := range state.specs {
			info = append(info, spec.ChildInfo{
				Id:   id,
				PID:  getPID(id),
				Type: state.specs.Type(id),
			})
		}
		request.ChildrenInfo = info
		actor.Send(call.Sender, request)
	}

	return true
}


func copyMap(src map[spec.CancelablePID]string) (dst map[spec.CancelablePID]string) {
	dst = make(map[spec.CancelablePID]string)
	for k, v := range src {
		dst[k] = v
	}
	return
}