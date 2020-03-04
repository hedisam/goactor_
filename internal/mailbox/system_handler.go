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
						Reason:   sysmsg.Kill,
						Relation: sysmsg.Linked,
					})
				case sysmsg.Normal:
					// for supervisor
					return true, msg
				}
			}
		}
	// if some actor/supervisor sends a Shutdown command they also should close the context's done channel
	case sysmsg.Shutdown:
		// todo: shutdown based on the shutdown value
		if m.Utils().TrapExit() {
			return true, msg
		} else {
			panic(msg)
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
