package actor

var myPID *PID

type registryMap map[string]*PID

type cmdRegister struct {
	name string
	pid  *PID
}
type cmdUnregister struct {
	name string
}
type cmdGet struct {
	name   string
	sender *PID
}

func init() {
	myPID = Spawn(registry)
}
