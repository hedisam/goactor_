package goactor

import (
	"fmt"
	"log"
)

const (
	OneForOneStrategy	SupervisorStrategy = iota
)

const (
	RestartAlways	int32 = iota
	RestartNever
	RestartTransient
)

const (
	Infinity	int32 = iota - 1
	Kill
)

type SupervisorStrategy	int32

type childSpecMap map[interface{}]ChildSpec

type registerChild struct {
	pid *PID
	name string
}

type ChildSpec struct {
	id	string
	start childSpecStart
	restart int32
	shutdown int32
	chType 	interface{}
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
		chType: "worker",
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
	specs := supervisor.Args()[0].(childSpecMap)
	//strategy := supervisor.Args()[1].(SupervisorStrategy)
	registry := map[*PID]string{}

	// todo: handle other strategies
	// default strategy is one-for-one
	supervisor.Recv(func(message interface{}) bool {
		switch msg := message.(type) {
		case registerChild:
			// register the linked child locally
			registry[msg.pid] = msg.name
		case NormalExit:
			// it's a normal termination
			name := registry[msg.who.pid]
			switch specs[name].restart {
			case RestartAlways:
				child := supervisor.SpawnLink(specs[name].start.actorFunc, specs[name].start.args)
				registry[child] = name
				Register(name, child)
			}
		case PanicExit:
			// it's an abnormal termination
			name := registry[msg.who.pid]
			switch specs[name].restart {
			case RestartAlways, RestartTransient:
				child := supervisor.SpawnLink(specs[name].start.actorFunc, specs[name].start.args)
				registry[child] = name
				Register(name, child)
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

