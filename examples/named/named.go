package main

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"time"
)

func main() {
	pid := actor.Spawn(echo)
	actor.Register("echo", pid)

	actor.SendNamed("echo", "hi")

	pid2 := actor.WhereIs("echo")
	actor.Send(pid2, "shutdown")

	time.Sleep(1 * time.Second)
}

func echo(actor *actor.Actor) {
	actor.Receive(func(message interface{}) (loop bool) {
		if message == "shutdown" {
			fmt.Println("[-] echo: shutting down")
			return false
		} else {
			fmt.Println("[+] echo:", message)
			return true
		}
	})
}