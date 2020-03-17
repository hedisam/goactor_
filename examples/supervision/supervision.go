package main

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/supervisor"
	"github.com/hedisam/goactor/supervisor/spec"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	supervisionTreeMain()

	wait()
}

func maxRestartsMain() {
	options := supervisor.NewOptions(supervisor.OneForOneStrategy, 2, 3)
	_, err := supervisor.Start(options, spec.NewWorkerSpec("panik", panik))
	if err != nil {
		log.Fatal(err)
	}

	// the first start doesn't count when counting max allowed restarts (a start is not a restart)
	for i := 0; i < 7; i++ {
		actor.SendNamed("panik", "hi panik, you wanna panic?")
		time.Sleep(10 * time.Millisecond)
	}
}

func panik(actor *actor.Actor) {
	fmt.Println("[+] panik started")
	actor.Receive(func(message interface{}) (loop bool) {
		fmt.Println("[!] panik received:", message)
		panic("just panic")
	})
}

func longRunningMain() {
	_, err := supervisor.Start(supervisor.OneForAllStrategyOption(),
		spec.NewWorkerSpec("#1", longRunning, "#1"),
		spec.NewWorkerSpec("#2", longRunning, "#2"),
	)
	if err != nil {
		log.Fatal(err)
	}

	actor.SendNamed("#1", "sleep")
	actor.SendNamed("#2", "panic")
}

func longRunning(actor *actor.Actor) {
	name := actor.Args()[0]
	fmt.Printf("[+] %v started\n", name)

	actor.Receive(func(message interface{}) (loop bool) {
		switch message {
		case "panic":
			panic("PANIC COMMAND")
		case "shutdown":
			return false
		case "sleep":
			select {
			case <-actor.Done():
				fmt.Printf("[!] %v is dead.\n", name)
				return false
			default:
				fmt.Printf("[!] %v sleeping for 1 sec\n", name)
				time.Sleep(1 * time.Second)
			}
			select {
			case <-actor.Done():
				fmt.Printf("[!] %v is dead.\n", name)
				return false
			default:
				fmt.Printf("[!] %v sleeping for 3 sec\n", name)
				time.Sleep(3 * time.Second)
			}
		}
		return true
	})
}

func simpleChildMain() {
	_, err := supervisor.Start(supervisor.OneForAllStrategyOption(),
		spec.NewWorkerSpec("#1", simpleChild, "#1").SetRestart(spec.RestartAlways),
		//spec.NewWorkerSpec("#2", simpleChild, "#2").SetShutdown(spec.ShutdownKill),
		spec.WorkerSpec{
			Id: "#3",
			Start: spec.WorkerStartSpec{
				ActorFunc: simpleChild,
				Args:      []interface{}{"#3"},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	actor.SendNamed("#1", "shutdown")
}

func simpleChild(actor *actor.Actor) {
	name := actor.Args()[0]
	fmt.Printf("[+] %v started\n", name)

	actor.Receive(func(message interface{}) (loop bool) {
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