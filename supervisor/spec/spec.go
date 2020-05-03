package spec

type Spec interface {
	ChildSpec() ChildSpec
}

type ChildSpec interface {
	ID() string
	Validate() error
	RestartValue() int32
	Start(supervisor linkSpawner) (error, CancelablePID)
}

type linkSpawner interface {
	SpawnLink() CancelablePID
}

type ChildType int32

type ChildInfo struct {
	Id   string
	PID  CancelablePID
	Type int32
}

const (
	WorkerActor int32 = iota
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