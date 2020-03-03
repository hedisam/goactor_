package main

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"os"
	"os/signal"
	"time"
)

func main() {
	var number int
	fmt.Printf("[+] Enter count: ")
	fmt.Scan(&number)
	fmt.Printf("[!] sending %d messages...\n", number)

	pid := actor.Spawn(counter, number)
	for i := 0; i < number; i++ {
		actor.Send(pid, "count me")
	}

	wait()
}

func counter(actor actor.Actor) {
	var number = actor.Context().Args()[0]
	count := 0
	now := time.Now()
	actor.Context().Recv(func(message interface{}) (loop bool) {
		count++
		if count == number {
			ellapsed := time.Since(now)
			fmt.Printf("[+] received %d messages in %v\n", count, ellapsed)
			return false
		}
		return true
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
