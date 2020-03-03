package process

import (
	"github.com/hedisam/goactor/actor"
)

func Register(name string, pid *actor.PID) {
	actor.Send(myPID, cmdRegister{name: name, pid: pid})
}

func Unregister(name string) {
	actor.Send(myPID, cmdUnregister{name: name})
}

func WhereIs(name string) (pid *actor.PID) {
	//future := goactor.NewParentActor()
	//goactor.Send(registryPID, cmdGet{name: name, sender: future.pid})
	//future.Recv(func(message interface{}) bool {
	//	if message == nil {
	//		pid = nil
	//	} else {
	//		pid = message.(*goactor.PID)
	//	}
	//	return false
	//})

	return
}

func registry(act actor.Actor) {
	repo := registryMap{}

	act.Context().Recv(func(message interface{}) (loop bool) {
		switch cmd := message.(type) {
		case cmdRegister:
			repo[cmd.name] = cmd.pid
		case cmdUnregister:
			repo[cmd.name] = nil
		case cmdGet:
			actor.Send(cmd.sender, repo[cmd.name])
		}
		return true
	})
}
