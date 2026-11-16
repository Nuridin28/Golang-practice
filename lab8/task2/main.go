// Task 2: Buffered vs unbuffered channels.
//
//	go run ./task2 run            2.1: run(...) with capacity 0, 1, 2, 3 and 10
//	go run ./task2 lenclose       len, cap and reading after close
//	go run ./task2 ab2            2.2: make(chan string, 2), two sends, then two receives
//	go run ./task2 ab1            2.2: the same with capacity 1: fatal error: deadlock
//	go run ./task2 fetch          participation: buffer 1 is enough for one result, not for two
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
		fmt.Println("usage: go run ./task2 run|lenclose|ab2|ab1|fetch")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "run":
		for _, size := range []int{0, 1, 2, 3, 10} {
			run(fmt.Sprintf("cap %2d", size), make(chan int, size))
		}
	case "lenclose":
		lenClose()
	case "ab2":
		ab(2)
	case "ab1":
		ab(1)
	case "fetch":
		fetchDemo()
	default:
		fmt.Println("unknown mode", os.Args[1])
		os.Exit(2)
	}
}

// run: the receiver is busy for 300 ms, then takes a value every 100 ms.
// The sender reports when each of its three sends returned.
func run(name string, ch chan int) {
	start := time.Now()
	since := func() time.Duration { return time.Since(start).Round(100 * time.Millisecond) }

	done := make(chan struct{})
	go func() {
		for i := 1; i <= 3; i++ {
			ch <- i
			fmt.Printf("%s: sent %d at %v\n", name, i, since())
		}
		close(done)
	}()

	time.Sleep(300 * time.Millisecond) // the receiver is busy at first
	for i := 0; i < 3; i++ {
		<-ch
		time.Sleep(100 * time.Millisecond) // and slow
	}
	<-done
}

func lenClose() {
	ch := make(chan int, 3)
	ch <- 1
	ch <- 2
	fmt.Println("len:", len(ch), "cap:", cap(ch))
	close(ch)
	v, ok := <-ch
	fmt.Println(v, ok)
	v, ok = <-ch
	fmt.Println(v, ok)
	v, ok = <-ch
	fmt.Println(v, ok)
}

// ab: two sends and two receives in ONE goroutine. That only works if both
// values fit into the buffer, because nobody else can receive.
func ab(size int) {
	ch := make(chan string, size)
	ch <- "a"
	ch <- "b" // size 1: the buffer is full and no other goroutine will ever receive
	fmt.Println(<-ch, <-ch)
}

// ---------------------------------------------------------------- participation

func slowQuery() int {
	time.Sleep(50 * time.Millisecond)
	return 42
}

// fetch: Week 8 fetchFixed, generalised to `sends` values into a channel of
// capacity `size`. The caller times out after 10 ms and never receives.
func fetch(size, sends int, timeout time.Duration) (int, error) {
	ch := make(chan int, size)
	go func() {
		v := slowQuery()
		for i := 0; i < sends; i++ {
			ch <- v // blocks once the buffer is full and nobody receives
		}
	}()
	select {
	case v := <-ch:
		return v, nil
	case <-time.After(timeout):
		return 0, errors.New("timeout")
	}
}

func fetchDemo() {
	for _, tc := range []struct{ size, sends int }{{0, 1}, {1, 1}, {1, 2}, {2, 2}} {
		before := runtime.NumGoroutine()
		for i := 0; i < 100; i++ {
			fetch(tc.size, tc.sends, 10*time.Millisecond)
		}
		time.Sleep(200 * time.Millisecond) // every slowQuery is finished by now
		fmt.Printf("cap %d, %d send(s): 100 timeouts -> %3d goroutines left behind\n",
			tc.size, tc.sends, runtime.NumGoroutine()-before)
	}
}
