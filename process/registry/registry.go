package process

import (
	"github.com/hedisam/goactor/actor"
)

var myPID *actor.PID

type registryMap map[string]*actor.PID

type cmdRegister struct {
	name string
	pid  *actor.PID
}
type cmdUnregister struct {
	name string
}
type cmdGet struct {
	name   string
	sender *actor.PID
}

func init() {
	myPID = actor.Spawn(registry)
}
