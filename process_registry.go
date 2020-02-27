package goactor

type registryMap map[string]*PID
var registryPID *PID
type registerCMD struct {
	name string
	pid *PID
}
type unregisterCMD struct {
	name string
}
type lookupCMD struct {
	name string
	sender *PID
}

func init() {
	registryPID = Spawn(registry)
}

func Register(name string, pid *PID) {
	Send(registryPID, registerCMD{name: name, pid: pid})
}

func Unregister(name string)  {
	Send(registryPID, unregisterCMD{name: name})
}

func WhereIs(name string) (pid *PID) {
	future := NewParentActor()
	Send(registryPID, lookupCMD{name: name, sender: future.pid})
	future.Recv(func(message interface{}) bool {
		if message == nil {
			pid = nil
		} else {
			pid = message.(*PID)
		}
		return false
	})

	return
}

func registry(actor *Actor) {
	repo := registryMap{}

	actor.Recv(func(message interface{}) bool {
		switch cmd := message.(type) {
		case registerCMD:
			repo[cmd.name] = cmd.pid
		case unregisterCMD:
			repo[cmd.name] = nil
		case lookupCMD:
			Send(cmd.sender, repo[cmd.name])
		}
		return true
	})
}