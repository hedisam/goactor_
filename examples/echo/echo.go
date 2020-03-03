package main

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"time"
)

func main() {
	pid := actor.Spawn(echo)

	actor.Send(pid, "Hello, this is Hidayat")
	actor.Send(pid, "shutdown")

	time.Sleep(2 * time.Second)
}

func echo(actor actor.Actor) {
	actor.Context().Recv(func(message interface{}) (loop bool) {
		if message == "shutdown" {
			fmt.Println("[-] echo: shutting down")
			return false
		} else {
			fmt.Println("[+] echo:", message)
			return true
		}
	})
}
