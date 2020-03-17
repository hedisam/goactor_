package main

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/internal/pid"
	"github.com/hedisam/goactor/sysmsg"
	"time"
)

type Message struct {
	Sender *pid.ProtectedPID
	Text string
}

func main() {
	parent, terminationHandler := actor.NewParentActor()
	defer terminationHandler()

	echoPID := actor.Spawn(echo)
	actor.Send(echoPID, Message{Sender: parent.Self(), Text: "Hey echo, send back this message"})
	actor.Send(echoPID, Message{Sender: parent.Self(), Text: "Another message from Hidayat"})

	parent.Context().RecvWithTimeout(1 * time.Second, func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case sysmsg.Timeout:
			future := actor.NewFutureActor()
			actor.Send(echoPID, Message{Sender: future.Self(), Text: "shutdown"})
			_, _ = future.Recv()
			fmt.Printf("[!] echo shutdown confirmed\n")
			return false
		case Message:
			fmt.Printf("[+] parent: %s\n", msg.Text)
		}
		return true
	})
}

func echo(a actor.Actor) {
	a.Context().Recv(func(message interface{}) (loop bool) {
		switch msg := message.(type) {
		case Message:
			actor.Send(msg.Sender, msg)
			if msg.Text == "shutdown" {
				return false
			}
			return true
		default:
			fmt.Println("[+] echo received unknown message:", message)
			return true
		}
	})
}
