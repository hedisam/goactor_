package spec

import "github.com/hedisam/goactor/internal/pid"

type Spec interface {
	ChildSpec() Spec
}
type ChildType int32

type ChildInfo struct {
	Id string
	PID *pid.ProtectedPID
	Type ChildType
}

const (
	TypeWorker ChildType = iota
	TypeSupervisor
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