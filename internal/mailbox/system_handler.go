package mailbox

import (
	"github.com/hedisam/goactor/sysmsg"
	"log"
)

// handleSystemMessage return true if the message should be passed to the user
func handleSystemMessage(m Mailbox, message interface{}) (bool, sysmsg.SystemMessage) {
	switch msg := message.(type) {
	case sysmsg.Exit:
		switch msg.Relation {
		case sysmsg.Monitored:
			return true, msg
		case sysmsg.Linked:
			if m.Utils().TrapExit() {
				return true, msg
			} else {
				switch msg.Reason {
				case sysmsg.Kill, sysmsg.Panic:
					panic(sysmsg.Exit{
						Who:      m.Utils().Self(),
						Parent:   msg.Who,
						Reason:   msg.Reason,
						Relation: sysmsg.Linked,
					})
				}
			}
		}
	// if some actor/supervisor sends a ShutdownFn command they also should close the context's done channel
	case sysmsg.Shutdown:
		// todo: shutdown based on the shutdown value
		if m.Utils().TrapExit() {
			return true, msg
		} else {
			panic(sysmsg.Exit{
				Who:      m.Utils().Self(),
				Parent:   msg.Parent,
				Reason:   sysmsg.Kill,
				Relation: sysmsg.Linked,
			})
		}
	case sysmsg.Monitor:
		if msg.Revert {
			m.Utils().DemonitorBy(msg.Parent)
		} else {
			m.Utils().MonitoredBy(msg.Parent)
		}
	case sysmsg.Link:
		if msg.Revert {
			m.Utils().Unlink(msg.To)
		} else {
			m.Utils().Link(msg.To)
		}
	default:
		log.Println("mailbox: unknown sys message", msg)
	}
	return false, nil
}

func checkContext(m Mailbox) {
	select {
	case <-m.Utils().ContextDone():
		// context's done channel is closed which means the actor has been shutdown by a supervisor
		// * the system message handler could've handled this already by panic(Exit) and since we have deferred
		// this function, we would catch that too.
		// * the user could've be doing some long running task and listened for the context's done channel then
		// returning false, meaning that the Shutdown command has never been processed by the system handler. in such a
		// case we should trigger panic(Exit) and that's the point of this code snippet.
		// * the user could've get the Shutdown command by setting trap exit and then returning,
		// in that case we're not gonna do anything.
		if r := recover(); r != nil {
			// already handled by the system handler. panic again since we've recovered from the previous one.
			// or it can be another panic, but doesn't matter since this actor has been declared as dead and respawned(?)
			// so it's not gonna be respawned two times(once because of being shutdown by the supervisor and another time
			// because of a new panic)
			panic(r)
		} else if m.Utils().TrapExit() {
			// handled by user
			return
		} else {
			// we don't have a ref to the parent supervisor
			panic(sysmsg.Exit{
				Who:      m.Utils().Self,
				Parent:   nil,
				Reason:   sysmsg.Kill,
				Relation: sysmsg.Linked,
			})
		}
	default:
		// context's done channel is not closed
	}
}
