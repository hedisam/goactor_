package actor

func Register(name string, pid *PID) {
	Send(myPID, cmdRegister{name: name, pid: pid})
}

func Unregister(name string) {
	Send(myPID, cmdUnregister{name: name})
}

func WhereIs(name string) (pid *PID) {
	future := NewFutureActor()
	Send(myPID, cmdGet{name: name, sender: future.Self()})
	result, _ := future.Recv()
	pid, _ = result.(*PID)
	return
}

func registry(act Actor) {
	repo := registryMap{}

	act.Context().Recv(func(message interface{}) (loop bool) {
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
