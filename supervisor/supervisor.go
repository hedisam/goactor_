package supervisor

import (
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/supervisor/ref"
	"github.com/hedisam/goactor/supervisor/spec"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type Init struct {sender *pid.ProtectedPID}

func Start(options Options, specs ...spec.Spec) (*ref.Ref, error) {
	specsMap, err := spec.ToMap(specs)
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
	actor.Send(suPID, Init{sender: future.Self()})
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
		case Init:
			err := state.init()
			actor.Send(msg.sender, err)
		case sysmsg.Exit:
			switch msg.Reason {
			case sysmsg.Panic:
				name, dead, found := state.registry.get(msg.Who.(pid.PID))
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
				name, dead, found := state.registry.get(msg.Who.(pid.PID))
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
			case sysmsg.SupMaxRestart:
				// a child supervisor just killed itself because one of its child reached
				// max restarts allowed in the same Period
				name, dead, found := state.registry.get(msg.Who.(pid.PID))
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
			}
		case sysmsg.Shutdown:
			// parent supervisor wants us to shutdown
			log.Println("supervisor: shutdown command from parent supervisor")
		default:
			log.Println("supervisor received unknown message:", msg)
		}
		return true
	})
}