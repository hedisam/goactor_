package actor

type processRegistry interface {
	Register(string, UserPID)
	Unregister(string)
	WhereIs(string) UserPID
}

var registry processRegistry

func SetProcessRegistry(r processRegistry) {
	registry = r
}