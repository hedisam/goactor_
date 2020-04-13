package actor

import (
	"github.com/hedisam/goactor/sysmsg"
	"sync/atomic"
	"time"
)

const (
	trapExitNo int32 = iota
	trapExitYes
)

type Context interface {
	SupervisorContext
	ReceiveWithTimeout(d time.Duration, handler func(message interface{}) (loop bool))
}

type Actor struct {
	Context
	trapExit        int32
	self            ClosablePID
	connectedActors *connectedActorsController
	mySupervisor    UserPID
}

func newActor(ctx Context, pid ClosablePID) *Actor {
	actor := &Actor{
		Context:         ctx,
		trapExit:        trapExitNo,
		self:            pid,
		connectedActors: newConnectedActorsController(),
	}
	return actor
}

func (a *Actor) trapExited() bool {
	if atomic.LoadInt32(&a.trapExit) == trapExitYes {
		return true
	}
	return false
}

func (a *Actor) Monitor(pid UserPID) {
	request := sysmsg.Monitor{Parent: a.self}
	pid.SendSystemMessage(request)
}

func (a *Actor) Demonitor(pid UserPID) {
	request := sysmsg.Monitor{Parent: a.self, Revert: true}
	pid.SendSystemMessage(request)
}

func (a *Actor) Link(pid UserPID) {
	// send a link request to the target Actor
	req := sysmsg.Link{To: a.self}
	pid.SendSystemMessage(req)

	// add the target to our linked actors list
	a.connectedActors.link(pid)
}

func (a *Actor) Unlink(pid UserPID) {
	// send an unlink request to the target Actor
	req := sysmsg.Link{To: a.self, Revert: true}
	pid.SendSystemMessage(req)

	// delete from linked actors list
	a.connectedActors.unlink(pid)
}

func (a *Actor) SpawnLink(fn Func, args ...interface{}) UserPID {
	pid := spawnLink(fn, a.self, args...)
	a.connectedActors.link(pid)
	return pid
}

func (a *Actor) SpawnMonitor(fn Func, args ...interface{}) UserPID {
	return spawnMonitor(fn, a.self, args...)
}

func (a *Actor) TrapExit(trapExit bool) {
	var trap = trapExitNo
	if trapExit {
		trap = trapExitYes
	}
	atomic.StoreInt32(&a.trapExit, trap)
}

func (a *Actor) Self() UserPID {
	return a.self
}

func (a *Actor) handleTermination() {
	// close Actor's mailbox done channel so it can't accept any further messages
	a.self.Dispose()

	// check if we got a panic or just a normal return
	switch r := recover().(type) {
	// a linked Actor terminated or got a sysmsg.Shutdown command by a supervisor
	// notify monitors and other linked actors
	case sysmsg.Exit:
		a.connectedActors.notify(&r)
	case sysmsg.Shutdown:
		// there's a case where user trap exit and receives the sysmsg.Shutdown msg then panics with the same msg
		exit := sysmsg.Exit{
			Who:    a.self,
			Parent: r.Parent,
			Reason: sysmsg.Reason{Type: sysmsg.Kill, Details: "shutdown cmd received from supervisor"},
		}
		a.connectedActors.notify(&exit)
	default:
		// something went wrong. notify monitors and linked actors with an Exit msg
		if r != nil {
			exit := sysmsg.Exit{
				Who:    a.self,
				Reason: sysmsg.Reason{Type: sysmsg.Panic, Details: r},
			}
			// check if the Actor is a supervisor, if so, then it had an unexpected panic and got no chances to
			// shutdown its children
			//shutdownChildren := a.actorType() == SupervisorActor
			a.connectedActors.notify(&exit)
			return
		}
		// it's a normal exit
		normal := sysmsg.Exit{
			Who: a.self,
			Reason: sysmsg.Reason{
				Type: sysmsg.Normal,
			},
		}
		a.connectedActors.notify(&normal)
	}
}