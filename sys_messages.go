package goactor

// SystemMessage represents a system message which could be just a notification or a command
type SystemMessage interface {
	sysMsg()
}

// exit commands are sent to linked actors to terminate them or they could avoid such terminations by setting trap_exit
// an actor that set trap_exit to true converts exit commands to a notify messages
// other exit messages act like a notification for monitor actors

// NormalExit describes a normal termination
type NormalExit struct {
	who *Actor
}
func (m NormalExit) sysMsg() {}
func (m NormalExit) Who() *Actor {
	return m.who
}

// PanicExit describes a termination happened by panic
type PanicExit struct {
	who *Actor
	reason string
}
func(m PanicExit) sysMsg() {}

// ExitCMD describes a command sent to a linked actor, making it terminate
type ExitCMD struct {
	becauseOf *Actor
	reason    string
}
func (c ExitCMD) sysMsg() {}

// KillExit describes a situation where the linked actor has to terminate because of another actor
type KillExit struct {
	who *Actor
	by	*Actor
	reason string
}
func (m KillExit) sysMsg() {}

// MonitorRequest is used to (de)monitor an already spawned actor
type MonitorRequest struct {
	who *Actor
	by 	*Actor
	// for un-monitoring set demonitor to true
	demonitor bool
}
func (r MonitorRequest) sysMsg() {}

// LinkRequest is used to (un)link to an already spawned actor
type LinkRequest struct {
	who 	*Actor
	to 		*Actor
	unlink 	bool
}
func (r LinkRequest) sysMsg() {}