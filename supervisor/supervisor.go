package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

// todo: implement supervisors shutdown

type Init struct {sender *pid.ProtectedPID}

func Start(options options, specs ...ChildSpec) (*pid.ProtectedPID, error) {
	specsMap, err := specsToMap(specs)
	if err != nil {
		return nil, err
	}

	err = checkOptions(&options)
	if err != nil {return nil, err}

	// spawn supervisor actor passing children specs data and the strategy as arguments
	suPID := actor.Spawn(supervisor, specsMap, options)
	actor.Register(options.name, suPID)

	// wait till all children are spawned
	future := actor.NewFutureActor()
	actor.Send(suPID, Init{sender: future.Self()})
	_, _ = future.Recv()

	return suPID, nil
}

func supervisor(supervisor actor.Actor) {
	// set trap exit since the supervisor is linked to its children
	supervisor.TrapExit(true)

	children := supervisor.Context().Args()[0].(childSpecMap)
	options := supervisor.Context().Args()[1].(options)

	registry := newRegistry(&options)

	shutdown := func(name string, _pid pid.PID) {
		registry.dead(_pid)
		actor.Send(pid.NewProtectedPID(_pid), sysmsg.Shutdown{
			Parent:   pid.ExtractPID(supervisor.Self()),
			Shutdown: children[name].Shutdown,
		})
		_pid.Shutdown()()
	}

	maxRestartsReached := func() {
		// shutdown all children and also the supervisor. restart the supervisor only if we have a child
		// with restart value set to RestartAlways
		// note: calling panic in supervisor should kill its children since they are linked but we're explicitly
		// shutting down each one to close children's context's done channel
		reg := copyMap(registry.aliveActors)
		for _pid, id := range reg {
			shutdown(id, _pid)
		}

		panic(sysmsg.Exit{
			Who:      pid.ExtractPID(supervisor.Self()),
			Parent:   nil,
			Reason:   sysmsg.SupMaxRestart,
			Relation: sysmsg.Linked,
		})
	}

	spawn := func(name string) {
		if registry.reachedMaxRestarts(name) {
			maxRestartsReached()
			return
		}

		child := supervisor.SpawnLink(children[name].Start.ActorFunc, children[name].Start.Args...)
		// register locally
		registry.put(pid.ExtractPID(child), name)
		// register globally
		actor.Register(name, child)
	}

	init := func() {
		for id := range children {
			spawn(id)
		}
	}

	handleOneForAll := func(name string) {
		reg := copyMap(registry.aliveActors)
		for _pid, id := range reg {
			_pid := _pid
			if id == name {
				// this actor already been terminated so no need to shut it down but we need to declare it as dead
				registry.dead(_pid)
			} else {
				shutdown(id, _pid)
			}
			spawn(id)
		}
	}

	handleRestForOne := func(name string) {
		log.Println("supervisor: rest_for_one strategy")
	}

	supervisor.Context().Recv(func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case Init:
			init()
			actor.Send(msg.sender, "ok")
		case sysmsg.Exit:
			switch msg.Reason {
			case sysmsg.Panic:
				name, dead, found := registry.get(msg.Who.(pid.PID))
				if dead || !found {
					return true
				}
				switch children[name].Restart {
				case RestartAlways, RestartTransient:
					switch options.strategy {
					case OneForOneStrategy:
						spawn(name)
					case OneForAllStrategy:
						handleOneForAll(name)
					case RestForOneStrategy:
						handleRestForOne(name)
					}
				}
			case sysmsg.Kill:
				// in result of sending a shutdown msg
				log.Println("supervisor: kill")
			case sysmsg.Normal:
				name, dead, found := registry.get(msg.Who.(pid.PID))
				if dead || !found {
					return true
				}
				if children[name].Restart == RestartAlways {
					switch options.strategy {
					case OneForOneStrategy:
						spawn(name)
					case OneForAllStrategy:
						handleOneForAll(name)
					case RestForOneStrategy:
						handleRestForOne(name)
					}
				}
			case sysmsg.SupMaxRestart:
				// a supervisor just killed itself because a child reaching max restarts allowed in the same period
				log.Println("supervisor:", msg.Reason)
			}
		default:
			log.Println("supervisor received unknown message:", msg)
		}
		return true
	})
}

func copyMap(src map[pid.PID]string) (dst map[pid.PID]string) {
	dst = make(map[pid.PID]string)
	for k, v := range src {
		dst[k] = v
	}
	return
}

// todo: we should register system processes on a different process registry
func checkOptions(options *options) error {
	if options.name == "" {
		return fmt.Errorf("invalid supervisor name: %s", options.name)
	} else if options.strategy < 0 || options.strategy > 2 {
		return fmt.Errorf("invalid strategy: %d", options.strategy)
	} else if options.period < 0 {
		return fmt.Errorf("invalid max seconds: %d", options.period)
	} else if options.maxRestarts < 0 {
		return fmt.Errorf("invalid max restarts: %d", options.maxRestarts)
	}

	return nil
}