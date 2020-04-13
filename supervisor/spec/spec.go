package spec

type Spec interface {
	ChildSpec() Spec
}

type ChildType int32

type ChildInfo struct {
	Id   string
	PID  CancelablePID
	Type ChildType
}

const (
	WorkerActor ChildType = iota
	SupervisorActor
)

const (
	RestartAlways int32 = iota
	RestartTransient
	RestartNever
)

const (
	ShutdownInfinity int32 = iota - 1 // -1
	ShutdownKill                      // 0
	// >= 1 as number of milliseconds
)