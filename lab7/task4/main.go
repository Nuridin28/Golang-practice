// Task 4: Explore the scheduler.
//
//	go build -o /tmp/sched ./task4
//	GOMAXPROCS=2 GODEBUG=schedtrace=500 /tmp/sched busy    8 busy goroutines on 2 P (as in the lesson)
//	GODEBUG=schedtrace=500 /tmp/sched busy                 ... on all cores
//	GOMAXPROCS=1 GODEBUG=schedtrace=500 /tmp/sched busy    ... on 1 P
//
//	/tmp/sched preempt                                     a busy loop can't freeze main
//	GODEBUG=asyncpreemptoff=1 /tmp/sched preempt           hangs forever (Ctrl+C)
//
//	/tmp/sched threads                                     time.Sleep vs a blocking system call
package main

import (
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: sched busy|preempt|threads")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "busy":
		busyGoroutines()
	case "preempt":
		preempt()
	case "threads":
		threads()
	default:
		fmt.Println("unknown mode", os.Args[1])
		os.Exit(2)
	}
}

var sink uint64

// spin does n iterations of pure CPU work and never blocks.
func spin(n int) {
	x := uint64(1)
	for i := 0; i < n; i++ {
		x = x*6364136223846793005 + 1442695040888963407
	}
	sink += x
}

// itersPerSecond measures how many spin iterations one core does in a second.
// It keeps the best of several tries: the M4 has fast performance cores and
// slower efficiency cores, and macOS may start us on either.
func itersPerSecond() int {
	const n = 20_000_000
	best := 0
	for try := 0; try < 10; try++ {
		start := time.Now()
		spin(n)
		best = max(best, int(float64(n)/time.Since(start).Seconds()))
	}
	return best
}

// busy does 2 seconds of CPU work, measured on an otherwise idle core.
//
// Not `for time.Since(start) < d {}`: that measures wall time. With 8 such
// goroutines on 2 P each one gets only 1/4 of a core, but they all still
// "finish" after 2 seconds, so the run would always take 2 s.
func busy(iters int) { spin(iters) }

// busyGoroutines: 8 goroutines, each with 2 seconds of CPU work.
func busyGoroutines() {
	fmt.Printf("NumCPU=%d GOMAXPROCS=%d\n", runtime.NumCPU(), runtime.GOMAXPROCS(0))
	twoSeconds := 2 * itersPerSecond()
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			busy(twoSeconds)
		}()
	}
	wg.Wait()
	fmt.Printf("8 x 2s of CPU work took %v\n", time.Since(start).Round(time.Millisecond))
}

// preempt: with one P, the busy loop gets the CPU as soon as main sleeps.
// main can only print if the runtime takes the CPU away from the loop.
func preempt() {
	runtime.GOMAXPROCS(1) // only one goroutine can run at a time
	go func() {
		for { // a busy loop with no function calls: no place to stop voluntarily
		}
	}()
	time.Sleep(10 * time.Millisecond)
	fmt.Println("main is still alive")
}

// threads: 100 goroutines wait 200 ms, once parked by the runtime, once blocked in the kernel.
func threads() {
	fmt.Printf("GOMAXPROCS=%d\n", runtime.GOMAXPROCS(0))
	fmt.Printf("%-24s threads created so far: %d\n", "start", threadCount())
	run("100 x time.Sleep", 100, func() { time.Sleep(200 * time.Millisecond) })
	run("100 x blocking syscall", 100, func() {
		// macOS has no syscall.Nanosleep. select(2) with no file descriptors and a
		// timeout is the classic "sleep in the kernel": the whole OS thread is blocked.
		tv := syscall.NsecToTimeval(int64(200 * time.Millisecond))
		_ = syscall.Select(0, nil, nil, nil, &tv)
	})
	run("100 x time.Sleep again", 100, func() { time.Sleep(200 * time.Millisecond) })
}

func run(name string, n int, f func()) {
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			f()
		}()
	}
	wg.Wait()
	fmt.Printf("%-24s threads created so far: %-4d (took %v)\n",
		name, threadCount(), time.Since(start).Round(time.Millisecond))
}

func threadCount() int { return pprof.Lookup("threadcreate").Count() }
