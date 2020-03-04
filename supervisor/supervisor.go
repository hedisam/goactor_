package supervisor

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

// todo: implement the max times a child could be restarted in a specific interval

const (
	// if a child process terminates, only that process is restarted
	OneForOneStrategy Strategy = iota

	// if a child process terminates, all other child processes are terminated
	// and then all of them (including the terminated one) are restarted.
	OneForAllStrategy

	// if a child process terminates, the terminated child process and
	// the rest of the children started after it, are terminated and restarted.
	RestForOneStrategy
)

type Strategy int32

func Start(strategy Strategy, specs ...*ChildSpec) (*pid.ProtectedPID, error) {
	specsMap, err := specsToMap(specs)
	if err != nil {
		return nil, err
	}

	// spawn supervisor actor passing children specs data and the strategy as arguments
	suPID := actor.Spawn(supervisor, specsMap, strategy)
	return suPID, nil
}

func supervisor(supervisor actor.Actor) {
	// set trap exit since the supervisor is linked to its children
	supervisor.TrapExit(true)

	registry := map[pid.PID]string{}
	children := supervisor.Context().Args()[0].(childSpecMap)
	strategy := supervisor.Context().Args()[1].(Strategy)

	spawn := func(name string) {
		child := supervisor.SpawnLink(children[name].start.actorFunc, children[name].start.args...)
		// register locally
		registry[pid.ExtractPID(child)] = name
		// register globally
		actor.Register(name, child)
	}

	for id := range children {
		spawn(id)
	}

	shutdown := func(name string, _pid pid.PID) {
		// todo: close the actor context [context.Context]
		actor.Send(pid.NewProtectedPID(_pid), sysmsg.Shutdown{
			Parent:   supervisor,
			Shutdown: children[name].shutdown,
		})
	}

	supervisor.Context().Recv(func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case sysmsg.Exit:
			switch msg.Reason {
			// todo: we should have a specific reason for "shutdown by supervisor"
			case sysmsg.Kill, sysmsg.Panic:
				name := registry[msg.Who.(pid.PID)]
				switch children[name].restart {
				case RestartAlways, RestartTransient:
					switch strategy {
					case OneForOneStrategy:
						spawn(name)
					case OneForAllStrategy:
						for _pid, name := range registry {
							_pid := _pid
							shutdown(name, _pid)
							spawn(name)
						}
					case RestForOneStrategy:
						panic("implement RestartForOneStrategy")
					}
				}
			case sysmsg.Normal:
				name := registry[msg.Who.(pid.PID)]
				if children[name].restart == RestartAlways {
					switch strategy {
					case OneForOneStrategy:
						spawn(name)
					case OneForAllStrategy:
						for _pid, name := range registry {
							_pid := _pid
							shutdown(name, _pid)
							spawn(name)
						}
					case RestForOneStrategy:
						panic("implement RestForOneStrategy")
					}
				}
			}
		default:
			log.Println("supervisor received unknown message:", msg)
		}
		return true
	})
}

func specsToMap(specs []*ChildSpec) (specsMap childSpecMap, err error) {
	if len(specs) == 0 {
		err = fmt.Errorf("empty childspec list")
		return
	}
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

		specsMap[s.id] = *s
	}
	return
}
