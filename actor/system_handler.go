package actor

import (
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

type systemHandler struct {
	actor *Actor
}

// HandleSystemMessage is called by the mailbox receiver for system messages
func (sysHandler *systemHandler) HandleSystemMessage(message interface{}) (bool, interface{}) {
	switch msg := message.(type) {
	case sysmsg.Exit:
		switch msg.Relation {
		case sysmsg.Monitored:
			return true, msg
		case sysmsg.Linked:
			if sysHandler.actor.trapExited() {
				return true, msg
			}
			switch msg.Reason.Type {
			case sysmsg.Kill, sysmsg.Panic:
				panic(sysmsg.Exit{
					Who:      sysHandler.actor.Self(),
					Parent:   msg.Who,
					Reason:   msg.Reason,
					Relation: sysmsg.Linked,
				})
			}
		}
	// supervisor sending shutdown command should also close the context's done channel
	case sysmsg.Shutdown:
		// todo: shutdown based on the shutdown value
		if sysHandler.actor.trapExited() {
			return true, msg
		}
		panic(sysmsg.Exit{
			Who:    sysHandler.actor.Self(),
			Parent: msg.Parent,
			Reason:   sysmsg.Reason{
				Type:    sysmsg.Kill,
				Details: "shutdown cmd received from supervisor",
			},
			Relation: sysmsg.Linked,
		})
	case sysmsg.Monitor:
		if msg.Revert {
			sysHandler.actor.connectedActors.demoniteredBy(msg.Parent)
		} else {
			sysHandler.actor.connectedActors.monitoredBy(msg.Parent)
		}
	case sysmsg.Link:
		if msg.Revert {
			sysHandler.actor.connectedActors.unlink(msg.To)
		} else {
			sysHandler.actor.connectedActors.link(msg.To)
		}
	default:
		log.Println("mailbox: unknown sys message", msg)
	}
	return false, nil
}

// CheckUnhandledShutdown is deferred by the mailbox receiver and handles unhandled shutdown commands sent by supervisor
func (sysHandler *systemHandler) CheckUnhandledShutdown() {
	select {
	case <-sysHandler.actor.Done():
		// context's done channel is closed which means the actor has been shutdown by sysHandler supervisor
		// * the system message handler could've handled this already by panic(Exit) and since we have deferred
		// this function, we would catch that too.
		// * the user could've be doing some long running task and listened for the context's done channel then
		// returning false, meaning that the Shutdown command has never been processed by the system handler. in such sysHandler
		// case we should trigger panic(Exit) and that's the point of this code snippet.
		// * the user could've get the Shutdown command by setting trap exit and then returning,
		// in that case we're not gonna do anything.
		if r := recover(); r != nil {
			// already handled by the system handler. panic again since we've recovered from the previous one.
			// or it can be another panic, but doesn't matter since this actor has been declared as dead and respawned(?)
			// so it's not gonna be respawned two times(once because of being shutdown by the supervisor and another time
			// because of sysHandler new panic)
			panic(r)
		} else if sysHandler.actor.trapExited() {
			// handled by user
			return
		} else {
			// we don't have sysHandler ref to the parent supervisor
			panic(sysmsg.Exit{
				Who:    sysHandler.actor.Self(),
				Parent: nil,
				Reason:   sysmsg.Reason{
					Type:    sysmsg.Kill,
					Details: "shutdown cmd received from supervisor",
				},
				Relation: sysmsg.Linked,
			})
		}
	default:
		// context's done channel is not closed
	}
}
