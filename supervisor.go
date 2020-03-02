package goactor

import (
	"fmt"
	"log"
)

const (
	// if a child process terminates, only that process is restarted
	OneForOneStrategy	SupervisorStrategy = iota

	// if a child process terminates, all other child processes are terminated
	// and then all of them (including the terminated one) are restarted.
	OneForAllStrategy

	// if a child process terminates, the terminated child process and
	// the rest of the children started after it, are terminated and restarted.
	RestForOneStrategy
)

const (
	RestartAlways	int32 = iota
	RestartTransient
	RestartNever
)

const (
	ShutdownInfinity	int32 = iota - 1	// -1
	ShutdownKill							// 0
									// >= 1 as number of milliseconds
)

type SupervisorStrategy	int32
type childSpecMap map[string]ChildSpec
type ChildType string

type registerChild struct {
	pid *PID
	name string
}

type ChildSpec struct {
	id	string
	start childSpecStart
	restart int32
	shutdown int32
	childType 	ChildType
}

type childSpecStart struct {
	actorFunc	ActorFunc
	args 	[]interface{}
}

func NewChildSpec(id string, fn ActorFunc, args ...interface{}) ChildSpec {
	return ChildSpec{
		id: id,
		start: childSpecStart{actorFunc: fn, args: args},
		restart: RestartTransient,
		shutdown: 5000,
		childType: "worker",
	}
}

func StartSupervisor(specs []ChildSpec, strategy SupervisorStrategy) (*PID, error) {
	specsMap, err := specsToMap(specs)
	if err != nil {return nil, err}

	// spawn supervisor actor passing children specs data and the strategy as arguments
	suPID := Spawn(supervisor, specsMap, strategy)
	supervisor := suPID.mailbox.getActor()
	// set trap exit since the supervisor is linked to its children
	supervisor.TrapExit(true)
	for _, spc := range specs {
		child := supervisor.SpawnLink(spc.start.actorFunc, spc.start.args)
		// locally
		register(suPID, child, spc.id)
		// globally
		Register(spc.id, child)
	}

	return suPID, nil
}

// register the child in the supervisor's local repo
func register(supervisor *PID, pid *PID, name string) {
	Send(supervisor, registerChild{pid: pid, name: name})
}

func supervisor(supervisor *Actor) {
	registry := map[*PID]string{}
	children := supervisor.Args()[0].(childSpecMap)
	strategy := supervisor.Args()[1].(SupervisorStrategy)

	reSpawn := func(name string) {
		child := supervisor.SpawnLink(children[name].start.actorFunc, children[name].start.args)
		registry[child] = name
		Register(name, child)
	}

	shutdown := func(name string, pid *PID) {
		// todo: consider child shutdown value
		// todo: close the actor context [context.Context]
		Send(pid, KillExit{who: pid.mailbox.getActor(), by: supervisor, reason: "shutdown by supervisor"})
	}

	supervisor.Recv(func(message interface{}) bool {
		switch msg := message.(type) {
		case registerChild:
			// register the linked child locally
			registry[msg.pid] = msg.name
		case NormalExit:
			// it's a normal termination
			name := registry[msg.who.pid]
			if children[name].restart == RestartAlways {
				switch strategy {
				case OneForOneStrategy:
					// just restart this actor
					reSpawn(name)
				case OneForAllStrategy:
					// shutdown and restart all children
					for pid, name := range registry {
						shutdown(name, pid)
						reSpawn(name)
					}
				case RestForOneStrategy:
					// restart this one and all actors that are started after this

				}
			}
		case PanicExit:
			// it's an abnormal termination
			name := registry[msg.who.pid]
			switch children[name].restart {
			case RestartAlways, RestartTransient:
				switch strategy {
				case OneForOneStrategy:
					reSpawn(name)
				case OneForAllStrategy:
					for pid, name := range registry {
						shutdown(name, pid)
						reSpawn(name)
					}
				}
			}
		default:
			log.Println("supervisor received unknown message:", msg)
		}
		return true
	})
}

func specsToMap(specs []ChildSpec) (specsMap childSpecMap, err error) {
	specsMap = childSpecMap{}
	for _, s := range specs {
		if s.id == "" {
			err = fmt.Errorf("childspec's id could not be empty")
			break
		} else if s.start.actorFunc == nil {
			err = fmt.Errorf("childspec's fn(ActorFunc) could not be nil")
			break
		} else if _, duplicate := specsMap[s.id]; duplicate {
			err = fmt.Errorf("duplicate childspec id")
			break
		}

		specsMap[s.id] = s
	}
	return
}

