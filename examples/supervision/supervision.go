package main

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/supervisor"
	"log"
	"os"
	"os/signal"
)

func main() {
	_, err := supervisor.Start(supervisor.OneForOneStrategy,
		supervisor.NewChildSpec("#1", child, "#1"),
		supervisor.NewChildSpec("#2", child, "#2"),
	)
	if err != nil {
		log.Fatal(err)
	}

	actor.SendNamed("#1", "shutdown")

	wait()
}

func child(actor actor.Actor) {
	name := actor.Context().Args()[0]
	fmt.Printf("[+] %v started\n", name)

	actor.Context().Recv(func(message interface{}) (loop bool) {
		switch message {
		case "shutdown":
			fmt.Printf("[-] %v shutting down\n", name)
			return false
		case "panic":
			panic("PANIC COMMAND")
		default:
			fmt.Printf("[!] %v received: %v\n", name, message)
			return true
		}
	})
}

func wait() {
	signalChan := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		fmt.Println("[!] CTRL+C")
		close(done)
	}()
	<-done
}