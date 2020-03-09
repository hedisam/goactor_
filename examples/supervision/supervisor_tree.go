package main

import (
	"fmt"
	"github.com/hedisam/goactor/actor"
	"github.com/hedisam/goactor/supervisor"
	"github.com/hedisam/goactor/supervisor/ref"
	"github.com/hedisam/goactor/supervisor/spec"
	"log"
	"time"
)

type worker struct {}
type mySupervisor struct {}

func supervisionTreeMain() {
	mySup := mySupervisor{}
	_, err := supervisor.Start(supervisor.OneForOneStrategyOption(), mySup)
	if err != nil {log.Fatal(err)}

	actor.SendNamed("worker", "panic")
	time.Sleep(1 * time.Millisecond)
	actor.SendNamed("worker", "panic")
	time.Sleep(2 * time.Millisecond)
	actor.SendNamed("worker", "your supervisor should've been restarted")
}

func (sup mySupervisor) ChildSpec() spec.Spec {
	return spec.NewSupervisorSpec(sup.StartLink, worker{})
}

func (sup mySupervisor) StartLink(specs ...spec.Spec) (*ref.Ref, error) {
	fmt.Println("[+] start_link invoked")
	return supervisor.Start(
		supervisor.NewOptions(supervisor.OneForOneStrategy, 1, 2),
		specs...)
}

func (w worker) ChildSpec() spec.Spec {
	return spec.NewWorkerSpec("worker", w.work)
}

func (w worker) work(actor actor.Actor) {
	fmt.Println("[!] work actor started")
	actor.Context().Recv(func(message interface{}) (loop bool) {
		switch message {
		case "panic":
			fmt.Println("[-] work actor received panic command")
			panic("work actor received panic command")
		default:
			fmt.Println("[!] work actor:", message)
			return true
		}
	})
}