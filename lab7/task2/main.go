// Task 2: First goroutines.
//
//	go run ./task2 say        the original program: main doesn't wait
//	go run ./task2 saysleep   the same with time.Sleep(100ms) in the loop (Tour of Go)
//	go run ./task2 sayfixed   2.1: fixed with sync.WaitGroup
//
//	go run ./task2 broken     2.2: the "two classic mistakes" program
//	go run -race ./task2 broken
//	go run ./task2 atomic     2.2 fix 1: wg.Add before go + atomic.Int64
//	go run ./task2 slice      2.2 fix 2: every goroutine writes its own element of a []int
//	go run -race ./task2 atomic && go run -race ./task2 slice
//
//	go run ./task2 panic      2.3: a goroutine panics, recover in main does nothing
//	go run ./task2 recover    2.3: the goroutine has its own deferred recover
//
//	go run ./task2 order      participation task: "worker 5 done" is printed first
//	./task2/run.sh            runs order 10 times with the default GOMAXPROCS and with GOMAXPROCS=1
package main

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./task2 say|saysleep|sayfixed|broken|atomic|slice|panic|recover|order")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "say":
		go say("goroutine", false)
		say("main", false)
	case "saysleep":
		go say("goroutine", true)
		say("main", true)
	case "sayfixed":
		sayFixed()
	case "broken":
		broken()
	case "atomic":
		fixedAtomic()
	case "slice":
		fixedSlice()
	case "panic":
		panicInGoroutine()
	case "recover":
		recoverInGoroutine()
	case "order":
		order()
	default:
		fmt.Println("unknown mode", os.Args[1])
		os.Exit(2)
	}
}

func say(msg string, sleep bool) {
	for i := 0; i < 3; i++ {
		if sleep {
			time.Sleep(100 * time.Millisecond)
		}
		fmt.Println(msg, i)
	}
}

// sayFixed: main waits for the goroutine, so both loops always finish.
// The order of the lines can still differ from run to run.
func sayFixed() {
	var wg sync.WaitGroup
	wg.Add(1) // count the goroutine BEFORE starting it
	go func() {
		defer wg.Done()
		say("goroutine", true)
	}()
	say("main", true)
	wg.Wait() // don't return from main until the goroutine is done
}

// broken has both classic mistakes: wg.Add inside the goroutine, and a data race on counter.
func broken() {
	var wg sync.WaitGroup
	counter := 0
	for i := 0; i < 1000; i++ {
		go func() {
			wg.Add(1) // WRONG: Wait() may run before this line
			defer wg.Done()
			counter++ // WRONG: 1000 goroutines change one variable
		}()
	}
	wg.Wait()
	fmt.Println("counter:", counter)
}

// fixedAtomic fixes both mistakes: Add before go, and an atomic counter.
func fixedAtomic() {
	var wg sync.WaitGroup
	var counter atomic.Int64
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter.Add(1) // read-add-write as one indivisible step
		}()
	}
	wg.Wait()
	fmt.Println("counter:", counter.Load())
}

// fixedSlice: no shared variable at all. Every goroutine owns one element;
// different elements are different memory locations, so there is no race.
// main reads them only after wg.Wait(), which "happens after" every Done().
func fixedSlice() {
	var wg sync.WaitGroup
	results := make([]int, 1000)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i]++
		}()
	}
	wg.Wait()
	counter := 0
	for _, r := range results {
		counter += r
	}
	fmt.Println("counter:", counter)
}

// panicInGoroutine: recover in main catches only panics of main's own goroutine.
func panicInGoroutine() {
	defer func() { fmt.Println("main recovered:", recover()) }() // never runs
	go func() {
		var m map[string]int
		m["x"] = 1 // panic inside a goroutine: the whole program dies
	}()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("main is still alive") // never printed
}

// recoverInGoroutine: the deferred recover is the first statement of the
// goroutine's own function, so it is on the stack of the goroutine that panics.
func recoverInGoroutine() {
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("goroutine recovered:", r)
			}
		}()
		var m map[string]int
		m["x"] = 1
	}()
	wg.Wait()

	// The same with a helper, as a real server would do it for every background task.
	wg.Add(1)
	safeGo(&wg, func() { panic("report generation failed") })
	wg.Wait()

	fmt.Println("main is still alive")
}

// safeGo starts f in a goroutine that cannot kill the program.
func safeGo(wg *sync.WaitGroup, f func()) {
	go func() {
		defer wg.Done()
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("safeGo recovered:", r)
			}
		}()
		f()
	}()
}

// order is the WaitGroup example from the lesson.
func order() {
	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Println("worker", i, "done")
		}()
	}
	wg.Wait()
	fmt.Println("all workers finished")
}
