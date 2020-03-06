package supervisor

import (
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

// todo: implement child supervisors
// todo: shutdown children in case of unhandled supervisor panics

type Init struct {sender *pid.ProtectedPID}

func Start(options Options, specs ...ChildSpec) (*pid.ProtectedPID, error) {
	specsMap, err := specsToMap(specs)
	if err != nil {
		return nil, err
	}

	err = options.checkOptions()
	if err != nil {return nil, err}

	// spawn supervisor actor passing specs specs data and options as arguments
	suPID := actor.Spawn(supervisor, specsMap, &options)
	// todo: register supervisors on a different process registry
	actor.Register(options.Name, suPID)

	// wait till all specs are spawned
	future := actor.NewFutureActor()
	actor.Send(suPID, Init{sender: future.Self()})
	_, _ = future.Recv()

	return suPID, nil
}

func supervisor(supervisor actor.Actor) {
	// set trap exit since the supervisor is linked to its specs
	supervisor.TrapExit(true)

	specs := supervisor.Context().Args()[0].(childSpecMap)
	options := supervisor.Context().Args()[1].(*Options)
	state := newState(specs, options, supervisor)

	//todo: unlink dead actors
	supervisor.Context().Recv(func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case Init:
			state.init()
			actor.Send(msg.sender, "ok")
		case sysmsg.Exit:
			switch msg.Reason {
			case sysmsg.Panic:
				name, dead, found := state.registry.get(msg.Who.(pid.PID))
				if dead || !found {
					return true
				}
				switch state.specs[name].Restart {
				case RestartAlways, RestartTransient:
					switch state.options.Strategy {
					case OneForOneStrategy:
						state.handleOneForOne(name, msg.Who.(pid.PID))
					case OneForAllStrategy:
						state.handleOneForAll(name)
					case RestForOneStrategy:
						state.handleRestForOne(name)
					}
				case RestartNever:
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
				switch state.specs[name].Restart {
				case RestartAlways:
					switch state.options.Strategy {
					case OneForOneStrategy:
						state.handleOneForOne(name, msg.Who.(pid.PID))
					case OneForAllStrategy:
						state.handleOneForAll(name)
					case RestForOneStrategy:
						state.handleRestForOne(name)
					}
				case RestartNever, RestartTransient:
					state.deadAndUnlink(msg.Who.(pid.PID))
				}
			case sysmsg.SupMaxRestart:
				// a supervisor just killed itself because a child reached max restarts allowed in the same Period
				log.Println("supervisor:", msg.Reason)
			}
		default:
			log.Println("supervisor received unknown message:", msg)
		}
		return true
	})
}