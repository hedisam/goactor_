package supervisor

import (
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/supervisor/spec"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type initMsg struct {sender spec.BasicPID}

func Start(options Options, specs ...spec.Spec) (*spec.SupRef, error) {
	specsMap, err := spec.ToMap(specs...)
	if err != nil {
		return nil, err
	}

	err = options.checkOptions()
	if err != nil {return nil, err}

	// spawn supervisor actor passing spec data and options as arguments
	suPID := actor.Spawn(supervisor, specsMap, &options)

	// wait till all spec are spawned
	future := actor.NewFutureActor()
	actor.Send(suPID, initMsg{sender: future.Self()})
	initErr, err := future.Receive()
	if err != nil {return nil, err}
	if initErr != nil {return nil, initErr.(error)}

	return &spec.SupRef{PID: suPID}, nil
}

func supervisor(supervisor *actor.Actor) {
	// set trap exit since the supervisor is linked to its children
	supervisor.TrapExit(true)

	specs := supervisor.Args()[0].(spec.SpecsMap)
	options := supervisor.Args()[1].(*Options)
	state := newState(specs, options, supervisor)

	supervisor.Receive(func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case initMsg:
			err := state.init()
			actor.Send(msg.sender, err)
		case sysmsg.Exit:
			switch msg.Reason.Type {
			case sysmsg.Panic, sysmsg.SupMaxRestart:
				name, dead, found := state.registry.id(msg.Who)
				if dead || !found {
					return true
				}
				switch state.specs.Restart(name) {
				case spec.RestartAlways, spec.RestartTransient:
					applyRestartStrategy(state, name, msg)
				case spec.RestartNever:
					state.deadAndUnlink(msg.Who)
				}
			case sysmsg.Kill:
				// in result of sending a shutdown msg
				log.Println("supervisor: kill")
			case sysmsg.Normal:
				name, dead, found := state.registry.id(msg.Who)
				if dead || !found {
					return true
				}
				switch state.specs.Restart(name) {
				case spec.RestartAlways:
					applyRestartStrategy(state, name, msg)
				case spec.RestartNever, spec.RestartTransient:
					state.deadAndUnlink(msg.Who)
				}
			}
		case sysmsg.Shutdown:
			// parent supervisor wants us to shutdown
			state.shutdownSupervisor(sysmsg.Reason{
				Type:    sysmsg.Kill,
				Details: "shutdown cmd received by parent supervisor",
			})
		case spec.Call:
			return state.handleCall(msg)
		default:
			log.Println("supervisor received unknown message:", msg)
		}
		return true
	})
}

func applyRestartStrategy(state *state, name string, msg sysmsg.Exit) {
	switch state.options.Strategy {
	case OneForOneStrategy:
		state.handleOneForOne(name, msg.Who)
	case OneForAllStrategy:
		state.handleOneForAll(name)
	case RestForOneStrategy:
		state.handleRestForOne(name)
	}
}