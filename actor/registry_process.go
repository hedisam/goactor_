package actor

import (
	"github.com/hedisam/goactor/internal/pid"
)

func Register(name string, pid *pid.ProtectedPID) {
	Send(myPID, cmdRegister{name: name, pid: pid})
}

func Unregister(name string) {
	Send(myPID, cmdUnregister{name: name})
}

func WhereIs(name string) (ppid *pid.ProtectedPID) {
	future := NewFutureActor()
	Send(myPID, cmdGet{name: name, sender: future.Self()})
	result, _ := future.Recv()
	ppid, _ = result.(*pid.ProtectedPID)
	return
}

func registry(act *Actor) {
	repo := registryMap{}

	act.Receive(func(message interface{}) (loop bool) {
		switch cmd := message.(type) {
		case cmdRegister:
			repo[cmd.name] = cmd.pid
		case cmdUnregister:
			repo[cmd.name] = nil
		case cmdGet:
			Send(cmd.sender, repo[cmd.name])
		}
		return true
	})
}
