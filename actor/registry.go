package actor

import "github.com/hedisam/goactor/internal/pid"

var myPID *pid.ProtectedPID

type registryMap map[string]*pid.ProtectedPID

type cmdRegister struct {
	name string
	pid  *pid.ProtectedPID
}
type cmdUnregister struct {
	name string
}
type cmdGet struct {
	name   string
	sender *pid.ProtectedPID
}

func init() {
	myPID = Spawn(registry)
}
