package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/supervisor/ref"
	"github.com/hedisam/goactor/supervisor/spec"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type initMsg struct {sender *pid.ProtectedPID}

func Start(options Options, specs ...spec.Spec) (*ref.Ref, error) {
	specsMap, err := spec.ToMap(specs...)
	if err != nil {
		return nil, err
	}

	err = options.checkOptions()
	if err != nil {return nil, err}

	// spawn supervisor actor passing spec data and options as arguments
	suPID := actor.Spawn(supervisor, specsMap, &options)
	pid.ExtractPID(suPID).ActorTypeFn()(actor.SupervisorActor)
	// todo: register supervisors on a different process registry
	actor.Register(options.Name, suPID)

	// wait till all spec are spawned
	future := actor.NewFutureActor()
	actor.Send(suPID, initMsg{sender: future.Self()})
	initErr, err := future.Recv()
	if err != nil {return nil, err}
	if initErr != nil {return nil, initErr.(error)}

	return &ref.Ref{PPID: suPID}, nil
}

func supervisor(supervisor actor.Actor) {
	// set trap exit since the supervisor is linked to its children
	supervisor.TrapExit(true)

	specs := supervisor.Context().Args()[0].(spec.SpecsMap)
	options := supervisor.Context().Args()[1].(*Options)
	state := newState(specs, options, supervisor)

	supervisor.Context().Recv(func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case initMsg:
			err := state.init()
			actor.Send(msg.sender, err)
		case sysmsg.Exit:
			switch msg.Reason {
			case sysmsg.Panic, sysmsg.SupMaxRestart:
				name, dead, found := state.registry.id(msg.Who.(pid.PID))
				if dead || !found {
					return true
				}
				switch state.specs.Restart(name) {
				case spec.RestartAlways, spec.RestartTransient:
					switch state.options.Strategy {
					case OneForOneStrategy:
						state.handleOneForOne(name, msg.Who.(pid.PID))
					case OneForAllStrategy:
						state.handleOneForAll(name)
					case RestForOneStrategy:
						state.handleRestForOne(name)
					}
				case spec.RestartNever:
					state.deadAndUnlink(msg.Who.(pid.PID))
				}
			case sysmsg.Kill:
				// in result of sending a shutdown msg
				log.Println("supervisor: kill")
			case sysmsg.Normal:
				name, dead, found := state.registry.id(msg.Who.(pid.PID))
				if dead || !found {
					return true
				}
				switch state.specs.Restart(name) {
				case spec.RestartAlways:
					switch state.options.Strategy {
					case OneForOneStrategy:
						state.handleOneForOne(name, msg.Who.(pid.PID))
					case OneForAllStrategy:
						state.handleOneForAll(name)
					case RestForOneStrategy:
						state.handleRestForOne(name)
					}
				case spec.RestartNever, spec.RestartTransient:
					state.deadAndUnlink(msg.Who.(pid.PID))
				}
			}
		case sysmsg.Shutdown:
			// parent supervisor wants us to shutdown
			// shutdown children
			reg := copyMap(state.registry.aliveActors)
			for _pid, id := range reg {
				state.shutdown(id, _pid)
			}
			panic(sysmsg.Exit{
				Who:      pid.ExtractPID(state.supervisor.Self()),
				Parent:   msg.Parent,
				Reason:   sysmsg.Kill,
				Relation: sysmsg.Linked,
			})
		case ref.Call:
			switch request := msg.Request.(type) {
			case ref.CountChildren:
				request.Specs = len(state.specs)
				request.Active = len(state.registry.aliveActors)
				for id, _ := range state.specs {
					if state.specs.Type(id) == spec.TypeSupervisor {
						request.Supervisors++
					} else {
						request.Workers++
					}
				}
				actor.Send(msg.Sender, request)
			case ref.DeleteChild:
				// check if a child exists with the specified id
				if _, exists := state.specs[request.Id]; !exists {
					actor.Send(msg.Sender, fmt.Errorf("child does not exists"))
					return true
				}
				// check if the child is running
				for _, id := range state.registry.aliveActors {
					if id == request.Id {
						// we can not delete a child that is running
						actor.Send(msg.Sender, fmt.Errorf("running child cannot be deleted"))
						return true
					}
				}
				// delete the child
				// note: if this supervisor gets restarted by a parent supervisor then the original child specs
				// will be used. (unless we update the parent with the new child specs)
				delete(state.specs, request.Id)
				actor.Send(msg.Sender, ref.OK{})
			case ref.RestartChild:
				// check if a child exists with the specified id
				if _, exists := state.specs[request.Id]; !exists {
					actor.Send(msg.Sender, fmt.Errorf("child does not exists"))
					return true
				}
				// check if the child is running
				for _, id := range state.registry.aliveActors {
					if id == request.Id {
						// we can not delete a child that is running
						actor.Send(msg.Sender, fmt.Errorf("running child cannot be deleted"))
						return true
					}
				}
				err := state.spawn(request.Id)
				if err != nil {
					actor.Send(msg.Sender, err)
					return true
				}
				actor.Send(msg.Sender, ref.OK{})
			case ref.StartChild:
				// check if the child spec is valid
				specMap, err := spec.ToMap(request.Spec)
				if err != nil {
					actor.Send(msg.Sender, err)
					return true
				}
				// check if we've already got a child spec with the same id
				var id string
				for id, _ = range specMap {}
				if _, exists := state.specs[id]; exists {
					actor.Send(msg.Sender, fmt.Errorf("a child spec already present with the same id"))
					return true
				}
				// add the child spec to the supervisor child spec map
				state.specs[id] = specMap[id]
				// start the child
				err = state.spawn(id)
				if err != nil {
					actor.Send(msg.Sender, err)
					return true
				}
				actor.Send(msg.Sender, ref.OK{})
			case ref.Stop:
				// todo: pass the reason, make sure it's is valid
				// shutdown children
				reg := copyMap(state.registry.aliveActors)
				for _pid, id := range reg {
					state.shutdown(id, _pid)
				}
				actor.Send(msg.Sender, ref.OK{})
				return false
			case ref.TerminateChild:
				// check if a child exists with the specified id
				if _, exists := state.specs[request.Id]; !exists {
					actor.Send(msg.Sender, fmt.Errorf("child does not exists"))
					return true
				}
				// check if the child is running
				for _pid, id := range state.registry.aliveActors {
					if id == request.Id {
						// found the alive child. shut it down
						state.shutdown(id, _pid)
						actor.Send(msg.Sender, request)
						return true
					}
				}
				// the child is not alive
				actor.Send(msg.Sender, fmt.Errorf("child already has been terminated"))
			case ref.WithChildren:
				getPID := func(id string) *pid.ProtectedPID {
					_pid := state.registry.pid(id)
					if _pid == nil {
						return nil
					}
					return pid.NewProtectedPID(_pid)
				}
				info := make([]spec.ChildInfo, 0, len(state.specs))
				for id, _ := range state.specs {
					info = append(info, spec.ChildInfo{
						Id:   id,
						PID:  getPID(id),
						Type: state.specs.Type(id),
					})
				}
				request.ChildrenInfo = info
				actor.Send(msg.Sender, request)
			}
		default:
			log.Println("supervisor received unknown message:", msg)
		}
		return true
	})
}