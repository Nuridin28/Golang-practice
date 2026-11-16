// Task 1: Channel semantics.
//
//	go run ./task1 ping          1.1: the sender sleeps 500 ms, the receiver waits on <-ch
//	go run ./task1 pingswap      1.1: the Sleep moved to the receiver (before <-ch)
//
//	go run ./task1 table         1.2: runs all 12 cells (each one in its own process) and prints a summary
//	go run ./task1 cell <name>   1.2: runs one cell, e.g. "nil-send", "closed-close" (see cells below)
//
//	go run ./task1 say           1.3: Week 8 say("goroutine"), main waits on a done channel instead of a WaitGroup
//
//	go run ./task1 stuck         participation: a goroutine blocked forever is NOT reported while main still has a timer
//	go run ./task1 stuckmain     participation: main blocked forever, but another goroutine sleeps in a loop: no report either
//	go run ./task1 leakprofile   participation: Go 1.26+ "goroutineleak" profile finds goroutines blocked on unreachable channels
package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./task1 ping|pingswap|table|cell <name>|say|stuck|stuckmain")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "ping":
		ping(false)
	case "pingswap":
		ping(true)
	case "table":
		table()
	case "cell":
		if len(os.Args) < 3 {
			fmt.Println("usage: go run ./task1 cell <name>")
			os.Exit(2)
		}
		runCell(os.Args[2])
	case "say":
		sayWithChannel()
	case "stuck":
		stuck()
	case "stuckmain":
		stuckMain()
	case "leakprofile":
		leakProfile()
	default:
		fmt.Println("unknown mode", os.Args[1])
		os.Exit(2)
	}
}

// ping: the example from Part 1. With sleepInReceiver the 500 ms sleep moves
// from the sender to the receiver, right before <-ch.
func ping(sleepInReceiver bool) {
	start := time.Now()
	since := func() time.Duration { return time.Since(start).Round(100 * time.Millisecond) }

	ch := make(chan string) // unbuffered
	go func() {
		if !sleepInReceiver {
			time.Sleep(500 * time.Millisecond)
		}
		fmt.Println("sender: about to send", since())
		ch <- "ping"
		fmt.Println("sender: send finished", since())
	}()

	fmt.Println("receiver: waiting... ", since())
	if sleepInReceiver {
		time.Sleep(500 * time.Millisecond)
	}
	msg := <-ch
	fmt.Println("receiver: got", msg, " ", since())
	// Give the sender a moment to print "send finished": once the hand-over is
	// done, main may otherwise return before the sender's Println runs.
	time.Sleep(10 * time.Millisecond)
}

// ---------------------------------------------------------------- 1.2: the table

// A cell is one operation on one kind of channel. The ones that end in a panic
// or a deadlock kill the process, so `table` runs every cell in a child process.
type cell struct {
	name string
	run  func()
}

var cells = []cell{
	// nil channel
	{"nil-send", func() {
		var ch chan int
		ch <- 1 // blocks forever
		fmt.Println("unreachable")
	}},
	{"nil-recv", func() {
		var ch chan int
		<-ch // blocks forever
		fmt.Println("unreachable")
	}},
	{"nil-close", func() {
		var ch chan int
		close(ch) // panic: close of nil channel
	}},
	{"nil-len", func() {
		var ch chan int
		fmt.Println("len:", len(ch), "cap:", cap(ch))
	}},

	// open channel: a partner goroutine makes the blocking operation finish
	{"open-send", func() {
		ch := make(chan int) // unbuffered: needs a receiver
		go func() {
			time.Sleep(100 * time.Millisecond)
			fmt.Println("receiver got", <-ch)
		}()
		t := time.Now()
		ch <- 1
		fmt.Println("send returned after", time.Since(t).Round(10*time.Millisecond))
		time.Sleep(10 * time.Millisecond)
	}},
	{"open-send-alone", func() { // the same without a receiver
		ch := make(chan int)
		ch <- 1
		fmt.Println("unreachable")
	}},
	{"open-recv", func() {
		ch := make(chan int)
		go func() {
			time.Sleep(100 * time.Millisecond)
			ch <- 42
		}()
		t := time.Now()
		v, ok := <-ch
		fmt.Println("got", v, ok, "after", time.Since(t).Round(10*time.Millisecond))
	}},
	{"open-recv-alone", func() {
		ch := make(chan int)
		<-ch
		fmt.Println("unreachable")
	}},
	{"open-close", func() {
		ch := make(chan int)
		close(ch)
		fmt.Println("closed fine")
	}},
	{"open-len", func() {
		ch := make(chan int, 3)
		ch <- 1
		ch <- 2
		fmt.Println("len:", len(ch), "cap:", cap(ch))
	}},

	// closed channel
	{"closed-send", func() {
		ch := make(chan int, 1)
		close(ch)
		ch <- 1 // panic: send on closed channel (even though the buffer has room)
	}},
	{"closed-recv", func() {
		ch := make(chan int, 3)
		ch <- 1
		ch <- 2
		close(ch)
		for i := 0; i < 4; i++ {
			v, ok := <-ch // 1 true, 2 true, then 0 false forever, without blocking
			fmt.Println(v, ok)
		}
	}},
	{"closed-close", func() {
		ch := make(chan int)
		close(ch)
		close(ch) // panic: close of closed channel
	}},
	{"closed-len", func() {
		ch := make(chan int, 3)
		ch <- 1
		ch <- 2
		close(ch)
		<-ch
		fmt.Println("len:", len(ch), "cap:", cap(ch)) // values left / buffer size
	}},
}

func runCell(name string) {
	for _, c := range cells {
		if c.name == name {
			c.run()
			return
		}
	}
	fmt.Println("unknown cell", name)
	os.Exit(2)
}

// table runs every cell as `<this binary> cell <name>` and classifies the result.
func table() {
	self, err := os.Executable()
	if err != nil {
		panic(err)
	}
	for _, c := range cells {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		var out bytes.Buffer
		cmd := exec.CommandContext(ctx, self, "cell", c.name)
		cmd.Stdout, cmd.Stderr = &out, &out
		err := cmd.Run()
		cancel()

		text := out.String()
		result := "OK"
		switch {
		case strings.Contains(text, "all goroutines are asleep"):
			result = "fatal error: deadlock"
		case strings.Contains(text, "panic:"):
			result = "panic"
		case err != nil:
			result = "error: " + err.Error()
		}
		fmt.Printf("%-16s %-22s %s\n", c.name, result, firstLines(text))
	}
}

// firstLines keeps the interesting line(s) of a cell's output on one line.
func firstLines(s string) string {
	var keep []string
	for _, l := range strings.Split(strings.TrimSpace(s), "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.HasPrefix(l, "goroutine ") && !strings.Contains(l, "[") {
			continue
		}
		if strings.HasPrefix(l, "main.") || strings.HasPrefix(l, "/") || strings.HasPrefix(l, "exit status") {
			continue
		}
		keep = append(keep, l)
		if len(keep) == 4 {
			break
		}
	}
	return strings.Join(keep, " | ")
}

// ---------------------------------------------------------------- 1.3: say

func say(msg string) {
	for i := 0; i < 3; i++ {
		time.Sleep(100 * time.Millisecond)
		fmt.Println(msg, i)
	}
}

// sayWithChannel: Week 8 sayFixed, but main waits on a channel, not a WaitGroup.
// The goroutine closes done when it is finished; <-done in main blocks until then.
func sayWithChannel() {
	done := make(chan struct{})
	go func() {
		defer close(done) // "I'm finished": also works if say panics and is recovered later
		say("goroutine")
	}()
	say("main")
	<-done
}

// ---------------------------------------------------------------- participation

// stuck: a goroutine is blocked on a channel forever, but main is still alive
// (it sleeps, like a server waits for the next request). The runtime only
// reports a deadlock when NO goroutine can ever run again — here main can,
// so the leak is silent.
func stuck() {
	ch := make(chan int)
	go func() {
		ch <- 1 // nobody will ever receive
	}()
	for i := 1; i <= 3; i++ {
		time.Sleep(200 * time.Millisecond)
		fmt.Printf("main still alive, goroutines: %d\n", runtime.NumGoroutine())
	}
	fmt.Println("exit without any deadlock error; 1 goroutine leaked")
}

// leakProfile: the deadlock detector only looks at "can ANY goroutine run?".
// The goroutineleak profile (Go 1.26+) asks a different question: is a
// goroutine blocked on a channel that no running goroutine can reach any
// more? It uses the garbage collector to find out. Here 3 goroutines block on
// channels that only they know about, and one blocks on a channel main still
// holds (that one is not reported: main could still send to it).
func leakProfile() {
	for i := 0; i < 3; i++ {
		go func() {
			ch := make(chan int) // only this goroutine references ch
			<-ch
		}()
	}
	reachable := make(chan int)
	go func() { <-reachable }()
	time.Sleep(50 * time.Millisecond)

	fmt.Println("goroutines:", runtime.NumGoroutine())
	if p := pprof.Lookup("goroutineleak"); p != nil {
		p.WriteTo(os.Stdout, 1)
	} else {
		fmt.Println("this Go version has no goroutineleak profile")
	}
	runtime.KeepAlive(reachable)
}

// stuckMain: main itself is blocked forever. But another goroutine sleeps in a
// loop (like a ticker or an http.Server's accept loop): the runtime can't
// prove that nobody will ever send to ch, so it waits. Press Ctrl+C.
func stuckMain() {
	go func() {
		for {
			time.Sleep(time.Second)
			fmt.Println("background goroutine: tick")
		}
	}()
	ch := make(chan int)
	fmt.Println("main: <-ch, blocked forever (Ctrl+C to stop)")
	<-ch
}
