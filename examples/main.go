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

const (
	zero int32 = iota - 1
	one
	two
	three
)

func main() {
	supervisor()

	wait()
}

func supervisor() {
	specs := []goactor.ChildSpec{
		goactor.NewChildSpec("panicee", panicee),
	}
	_, err := goactor.StartSupervisor(specs, goactor.OneForOneStrategy)
	if err != nil {
		log.Fatal(err)
	}
	// make panicee to panic
	goactor.SendNamed("panicee", Panic{payload: 0})

	time.Sleep(10 * time.Millisecond)
	goactor.SendNamed("panicee", "Hello panicee")
	goactor.SendNamed("panicee", "Hello panicee")
	goactor.SendNamed("panicee", "Hello panicee")
	// make panicee to shutdown normally
	goactor.SendNamed("panicee", Shutdown{})
}

func spawnMany() {
	count := inputCount()
	fmt.Println("[+] spawning", count, "actors")
	now := time.Now()
	for i := 0; i < count; i++ {
		goactor.Spawn(echo)
	}
	fmt.Println("[!] elapsed:", time.Since(now))
	wait()
}

func echoWithPauseMain() {
	pid := goactor.Spawn(echo)
	fmt.Println("echo spawned")

	var cmd string = ""
	for cmd != "shutdown" {
		fmt.Println("Enter msg: ")
		_, err := fmt.Scanf("%s", &cmd)
		if err != nil {
			log.Println(err)
			return
		}
		goactor.Send(pid, cmd)
	}
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

	parent.RecvWithTimeout(2 * time.Second, func(message interface{}) bool {
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
	log.Println("[+] panicee started")
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
	actor.RecvWithTimeout(1 * time.Second, func(message interface{}) bool {
		switch message.(type) {
		case goactor.TimeoutMessage:
			log.Println("counter: timeout")
			return false
		default:
		}
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
			fmt.Println("[!] echo is shutting down")
			return false
		default:
			fmt.Println("[!] echo:", message)
			//msg := message.(Message)
			//goactor.Send(msg.Origin, msg.Text)
			return true
		}
	})
}

func inputCount() int {
	fmt.Print("Enter count: ")
	var count int
	fmt.Scanf("%d", &count)
	return count
}

func wait() {
	fmt.Println("[!] Press ENTER to exit")
	var s string
	fmt.Scanln(&s)
}