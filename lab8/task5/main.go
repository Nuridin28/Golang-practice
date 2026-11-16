// Task 5: select.
//
//	go run ./task5 random          5.1: a and b both have 1000 values ready, 1000 selects
//	go run ./task5 prefer          5.1: the same, but a is preferred (two selects, the first with default)
//	go run ./task5 timeout         the timeout example + the non-blocking send (queue full, dropped)
//	go run ./task5 fetch           5.2: fetchWithTimeout, NumGoroutine before/after (no leak)
//	go run ./task5 merge           5.3: merge(a, b) with nil channels
//	go run ./task5 mergenonil      5.3: merge without `a = nil`: counts the useless zero values
//	go run ./task5 pingpong        participation: two players pass a ball for 100 ms
//
// Tests: go test -race -v ./task5
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./task5 random|prefer|timeout|fetch|merge|mergenonil|pingpong")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "random":
		fmt.Println("both ready, 1000 selects:", randomChoice(1000))
	case "prefer":
		fmt.Println("prefer a, 1000 selects:", preferA(1000))
	case "timeout":
		timeoutDemo()
	case "fetch":
		fetchDemo()
	case "merge":
		for v := range merge(source(1, 3, 10*time.Millisecond), source(100, 4, 7*time.Millisecond)) {
			fmt.Print(v, " ")
		}
		fmt.Println()
	case "mergenonil":
		for i := 0; i < 5; i++ {
			vals, zeros := mergeNoNil(source(1, 3, 0), source(100, 3, 2*time.Millisecond))
			fmt.Printf("run %d: values %v, useless zero values from a closed channel: %d\n", i+1, vals, zeros)
		}
	case "pingpong":
		pingPong(100 * time.Millisecond)
	default:
		fmt.Println("unknown mode", os.Args[1])
		os.Exit(2)
	}
}

// filled returns a buffered channel that already holds n copies of v.
func filled(v string, n int) chan string {
	ch := make(chan string, n)
	for i := 0; i < n; i++ {
		ch <- v
	}
	return ch
}

// randomChoice: both channels are always ready, so select picks uniformly at random.
func randomChoice(n int) map[string]int {
	a, b := filled("a", n), filled("b", n)
	count := map[string]int{}
	for i := 0; i < n; i++ {
		select {
		case v := <-a:
			count[v]++
		case v := <-b:
			count[v]++
		}
	}
	return count
}

// preferA: the first select only looks at a; default means "a is not ready".
// Only then do we wait on both. Result: b is used only when a is empty.
func preferA(n int) map[string]int {
	a, b := filled("a", n/2), filled("b", n) // a runs dry halfway, to show that b still works
	count := map[string]int{}
	for i := 0; i < n; i++ {
		select {
		case v := <-a:
			count[v]++
			continue
		default:
		}
		select {
		case v := <-a:
			count[v]++
		case v := <-b:
			count[v]++
		}
	}
	return count
}

func timeoutDemo() {
	slow := make(chan string, 1) // buffer 1: the sender can finish after we stop waiting
	go func() {
		time.Sleep(200 * time.Millisecond)
		slow <- "answer"
	}()
	select {
	case r := <-slow:
		fmt.Println("got", r)
	case <-time.After(100 * time.Millisecond):
		fmt.Println("timeout after 100ms")
	}

	queue := make(chan int, 2)
	for i := 1; i <= 4; i++ {
		select {
		case queue <- i:
			fmt.Println("queued", i)
		default:
			fmt.Println("queue full, dropped", i)
		}
	}
}

// ---------------------------------------------------------------- 5.2

var ErrTimeout = errors.New("timeout")

// fetchWithTimeout waits for slow() at most `timeout`.
//
// No leak: the result channel has a buffer of one, so the goroutine's send
// always succeeds, even if nobody receives any more, and the goroutine ends
// when slow() returns. (The work itself is not stopped — for that, slow needs a
// done channel or a context, see Task 6 and Week 8 Task 6.)
func fetchWithTimeout(slow func() int, timeout time.Duration) (int, error) {
	ch := make(chan int, 1)
	go func() { ch <- slow() }()

	t := time.NewTimer(timeout) // NewTimer + Stop instead of time.After: nothing waits for 'timeout' after we return
	defer t.Stop()
	select {
	case v := <-ch:
		return v, nil
	case <-t.C:
		return 0, ErrTimeout
	}
}

func fetchDemo() {
	slow := func() int { time.Sleep(50 * time.Millisecond); return 42 }
	before := runtime.NumGoroutine()

	v, err := fetchWithTimeout(slow, 100*time.Millisecond)
	fmt.Println("timeout 100ms:", v, err)
	v, err = fetchWithTimeout(slow, 10*time.Millisecond)
	fmt.Println("timeout 10ms: ", v, err)

	for i := 0; i < 1000; i++ {
		fetchWithTimeout(slow, time.Millisecond)
	}
	fmt.Println("right after 1000 timeouts:  ", runtime.NumGoroutine()-before, "extra goroutines (still inside slow)")
	time.Sleep(100 * time.Millisecond)
	fmt.Println("100ms later (slow is done):", runtime.NumGoroutine()-before, "extra goroutines")
}

// ---------------------------------------------------------------- 5.3

// source sends first, first+1, ... (n values), pausing between them, then closes.
func source(first, n int, pause time.Duration) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; i < n; i++ {
			if pause > 0 {
				time.Sleep(pause)
			}
			out <- first + i
		}
	}()
	return out
}

// merge reads from a and b until BOTH are closed.
func merge(a, b <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for a != nil || b != nil {
			select {
			case v, ok := <-a:
				if !ok {
					a = nil // a nil channel blocks forever: this case is now off
					continue
				}
				out <- v
			case v, ok := <-b:
				if !ok {
					b = nil
					continue
				}
				out <- v
			}
		}
	}()
	return out
}

// mergeNoNil is merge without the `= nil` lines (and it collects instead of
// forwarding, to keep the experiment simple). After a is closed, `<-a` is
// always ready and returns 0, false immediately, so select keeps picking it.
// We count those useless wake-ups. The loop needs another way to end: it stops
// when both channels have reported closed at least once.
func mergeNoNil(a, b <-chan int) (vals []int, zeros int) {
	aClosed, bClosed := false, false
	for !aClosed || !bClosed {
		select {
		case v, ok := <-a:
			if !ok {
				aClosed = true
				zeros++
				continue
			}
			vals = append(vals, v)
		case v, ok := <-b:
			if !ok {
				bClosed = true
				continue
			}
			vals = append(vals, v)
		}
	}
	return vals, zeros
}

// ---------------------------------------------------------------- participation

// pingPong: two players pass a ball over ONE unbuffered channel and add 1 each
// time. After d main takes the ball out of play and stops both players.
func pingPong(d time.Duration) {
	before := runtime.NumGoroutine()
	table := make(chan int)
	done := make(chan struct{})
	finished := make(chan struct{})

	player := func() {
		defer func() { finished <- struct{}{} }()
		for {
			select {
			case ball := <-table:
				ball++
				select {
				case table <- ball: // pass it to the other player...
				case <-done: // ...unless the game is over
					return
				}
			case <-done:
				return
			}
		}
	}
	go player()
	go player()

	table <- 0 // serve
	time.Sleep(d)
	ball := <-table // main catches the ball: nobody can hit it any more
	close(done)     // wakes BOTH players at once
	<-finished
	<-finished

	fmt.Printf("%v: %d hits (%.0f ns per hit)\n", d, ball, float64(d.Nanoseconds())/float64(ball))
	fmt.Println("goroutines before:", before, "after:", runtime.NumGoroutine())
}
