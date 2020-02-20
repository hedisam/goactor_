package main

import (
	"fmt"
	"github.com/hedisam/goactor"
	"log"
	"time"
)

type Message struct {
	Text string
	Origin *goactor.PID
	Payload interface{}
}

type Panic struct {payload int}
type Shutdown struct {}

var mCount int

func main() {
	monitoredMain()

	//wait()
}

func monitoredMain() {
	parent := goactor.NewParentActor()
	childes := make(map[*goactor.PID]int)
	for i := 0; i < 3; i++ {
		pid := parent.SpawnMonitor(monitored)
		childes[pid] = i + 1
		if i == 1 {
			parent.DeMonitor(pid)
		}
	}

	parent.After(2 * time.Second)
	parent.Recv(func(message interface{}) bool {
		switch msg := message.(type) {
		case goactor.TimeoutMessage:
			fmt.Println("parent timeout: exiting...")
			return false
		case goactor.NormalExit:
			fmt.Printf("[!] parent received NormalExit by: %v, actor #%d\n", msg.Who(), childes[msg.Who().Self()])
			return true
		default:
			fmt.Println("[!] parent received unknown message:", message)
			return true
		}
	})

	//parent.After(1 * time.Second)
	parent.RecvWithTimeout(1 * time.Second, func(message interface{}) bool {
		log.Println("hi")
		return false
	})
}

func monitored(actor *goactor.Actor) {
	fmt.Println("[+] actor spawned with id:", actor.Self().ID())
	actor.After(1 * time.Second)
	actor.Recv(func(message interface{}) bool {
		switch msg := message.(type) {
		case goactor.TimeoutMessage:
			fmt.Println("[!] actor timeout, id:", actor.Self().ID())
		default:
			log.Printf("[!] unknown message in actor %s: %v", actor.Self().ID(), msg)
		}
		return false
	})
}

func panicMain() {
	parent := goactor.NewParentActor()
	pid := goactor.Spawn(panicer)
	goactor.Send(pid, Message{Origin: parent.Self(), Text: "spawn the panicee"})
	parent.Monitor(pid)

	var pid2 *goactor.PID
	parent.Recv(func(message interface{}) bool {
		switch msg := message.(type) {
		case *goactor.PID:
			pid2 = msg
			log.Println("got the panicee pid")
		default:
			log.Println("unknown message from panicer")
		}
		return false
	})

	log.Println("sending commands to panicee :)")
	goactor.Send(pid2, Panic{0})
	time.Sleep(1 * time.Second)
	goactor.Send(pid, "Hello")

	parent.RecvWithTimeout(500 * time.Millisecond, func(message interface{}) bool {
		log.Println("parent: ", message)
		return false
	})
}

func panicer(actor *goactor.Actor) {
	actor.TrapExit(false)
	actor.Recv(func(message interface{}) bool {
		switch msg :=  message.(type) {
		case Message:
			log.Println("panicer:", msg.Text)
			pid := actor.SpawnLink(panicee)
			goactor.Send(msg.Origin, pid)
			return true
		default:
			log.Println("panicer:", msg)
			return true
		}
	})
}

func panicee(actor *goactor.Actor) {
	actor.Recv(func(message interface{}) bool {
		switch msg := message.(type) {
		case Shutdown:
			log.Println("panicee actor: shutdown")
			return false
		case Panic:
			log.Println("panicee actor: panic")
			//panic("panicee got a panic message")
			fmt.Println("panicee div:", 1/msg.payload)
			return false
		default:
			log.Println("panicee actor:", msg)
			return true
		}
	})
}

func counterMain(count int) {
	if count <= 0 {
		fmt.Print("Enter count: ")
		fmt.Scan(&mCount)
	} else {
		mCount = count
	}
	parent := goactor.NewParentActor()
	msg := Message{Origin: parent.Self()}
	pid := goactor.Spawn(counter)

	fmt.Printf("[!] sending #%d messages...\n", mCount)
	for i := 0; i <= mCount; i++ {
		goactor.Send(pid, msg)
	}

	parent.Recv(func(message interface{}) bool {
		return false
	})
}

func counter(actor *goactor.Actor) {
	i := 0
	var now time.Time
	actor.Recv(func(message interface{}) bool {
		if i == 0 {
			now = time.Now()
		} else if i == mCount {
			fmt.Printf("[+] receiving #%d messages took: %v\n", i, time.Since(now))
			msg := message.(Message)
			goactor.Send(msg.Origin, "ok")
			return false
		}
		i++
		//fmt.Println("[+] counter:", message)
		return true
	})
}

func sumStarter(actor *goactor.Actor) {
	actor.Recv(func(message interface{}) bool {
		switch msg := message.(type) {
		case Message:
			pid := goactor.Spawn(sum)
			goactor.Send(msg.Origin, pid)
			log.Println("sumStarter: sum spawned")
			return true
		case Shutdown:
			log.Println("sumStarter: Shutdown")
			return false
		default:
			return false
		}
	})
}

func sum(actor *goactor.Actor) {
	sum := 0
	actor.Recv(func(message interface{}) bool {
		switch msg := message.(type) {
		case Shutdown:
			log.Println("sum: shutdown")
			return false
		case Message:
			sum += msg.Payload.(int)
			log.Println("sum:", sum)
			return true
		default:
			return true
		}
	})
}

func echo(actor *goactor.Actor) {
	actor.Recv(func(message interface{}) bool {
		switch message {
		case "shutdown":
			return false
		default:
			log.Println("echo: ", message)
			msg := message.(Message)
			goactor.Send(msg.Origin, msg.Text)
			return true
		}
	})
}

func wait() {
	fmt.Println("[!] Press ENTER to exit")
	var s string
	fmt.Scanln(&s)
}