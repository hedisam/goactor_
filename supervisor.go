package goactor

import "fmt"

const (
	OneForOneStrategy	int32 = iota
)

type SupervisorStrategy	int32

type childSpecMap map[interface{}]ChildSpec

type registerChild struct {
	pid *PID
	name interface{}
}

func (r registerChild) sysMsg() {

}

type ChildSpec struct {
	id	interface{}
	start childSpecStart
	restart interface{}
	shutdown interface{}
	chType 	interface{}
}

type childSpecStart struct {
	actorFunc	ActorFunc
	args 	[]interface{}
}

func NewChildSpec(id interface{}, fn ActorFunc, args ...interface{}) *ChildSpec {
	return &ChildSpec{
		id: id,
		start: childSpecStart{actorFunc: fn, args: args},
		restart: "always",
		shutdown: 5000,
		chType: "worker",
	}
}

func StartSupervisor(specs []ChildSpec, strategy SupervisorStrategy) (*PID, error) {
	specsMap, err := specsToMap(specs)
	if err != nil {return nil, err}

	// spawn supervisor actor passing children specs data and the strategy as arguments
	pid := Spawn(supervisor, specsMap, strategy)
	supervisor := pid.mailbox.getActor()
	// set trap exit since the supervisor is linked to its children
	supervisor.TrapExit(true)
	for _, spc := range specs {
		// todo: register the linked actor
		child := supervisor.SpawnLink(spc.start.actorFunc, spc.start.args)
		register(pid, child, spc.id)
	}

	return pid, nil
}

func register(supervisor *PID, pid *PID, name interface{}) {
	sendSystem(supervisor, registerChild{pid: pid, name: name})
}

func supervisor(actor *Actor) {
	specs := actor.Args()[0].(childSpecMap)
	strategy := actor.Args()[1].(SupervisorStrategy)
	registry := map[interface{}]*PID{}

	actor.Recv(func(message interface{}) bool {
		switch msg := message.(type) {
		case registerChild:
			// register the linked child locally
			registry[msg.name] = msg.pid
		case NormalExit:
			// it's a normal termination

		case ExitCMD:
			// it's an abnormal termination

		}
		fmt.Println(specs, strategy)
		return true
	})
}

func specsToMap(specs []ChildSpec) (specsMap childSpecMap, err error) {
	specsMap = childSpecMap{}
	for _, s := range specs {
		if s.id == nil {
			err = fmt.Errorf("childspec's id could not be nil")
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

