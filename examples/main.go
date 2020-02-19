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

var count int

func main() {
	monitoredMain()

	wait()
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

	parent.RecvWithTimeout(4*time.Second, func(message interface{}) bool {
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

	time.AfterFunc(4 * time.Second, func() {
		goactor.Send(parent.Self(), "Hi")
	})

	parent.Recv(func(message interface{}) bool {
		log.Println("got:", message)
		return false
	})
}

func monitored(actor *goactor.Actor) {
	fmt.Println("[+] actor spawned with id:", actor.Self().ID())
	actor.RecvWithTimeout(1 * time.Second, func(message interface{}) bool {
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

	parent.Recv(func(message interface{}) bool {
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

func counterMain() {
	fmt.Print("Enter count: ")
	fmt.Scanf("%d", &count)
	fmt.Printf("\nSending %d messages...\n", count)

	parent := goactor.NewParentActor()
	pid := goactor.Spawn(counter)

	goactor.Send(pid, Message{Origin: parent.Self()})
	now := time.Now()
	for i:=0; i < count; i++ {
		goactor.Send(pid, "count")
	}

	fmt.Println("sent:", time.Since(now))

	parent.Recv(func(message interface{}) bool {
		switch message {
		case "finish":
			fmt.Println("time elapsed:", time.Since(now))
			return false
		default:
			return true
		}
	})
}

func counter(actor *goactor.Actor) {
	//format := ".txt"
	//path := "C:\\Users\\Hidayat\\Desktop\\test"
	var parent *goactor.PID
	i := 0
	actor.Recv(func(message interface{}) bool {
		switch msg := message.(type) {
		case Message:
			parent = msg.Origin
			return true
		default:
			i++
			//if i % 400000 == 0 {
			//	log.Println("reading a file...")
			//	data, err := ioutil.ReadFile(path + format)
			//	if err != nil {log.Println("reading err: ", err)}
			//	log.Println("writing the content to another file")
			//	err = ioutil.WriteFile(fmt.Sprintf("%s%d%s", path, i, format), data, 0644)
			//	if err != nil {log.Println("writing err:", err)}
			//}
			if i == count {
				fmt.Printf("received messages: #%d\n", i)
				goactor.Send(parent, "finish")
				return false
			}
			return true
		}
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
	var s string
	_, err := fmt.Scanln(&s)
	if err != nil {
		log.Println("wait err:", err)
	}
}