// Task 3: How much does a goroutine cost?
//
//	go run ./task3                 100 000 goroutines waiting on a channel
//	go run ./task3 -n 1000000      a million
//	go test -bench . -benchmem ./task3
package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	n := flag.Int("n", 100_000, "number of goroutines")
	flag.Parse()

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	stop := make(chan struct{})
	var ready, wg sync.WaitGroup // ready: started and waiting; wg: finished
	start := time.Now()
	for i := 0; i < *n; i++ {
		wg.Add(1)
		ready.Add(1)
		go func() {
			defer wg.Done()
			ready.Done()
			<-stop // wait until main closes the channel
		}()
	}
	ready.Wait() // every goroutine has really started, not just been created
	startTime := time.Since(start)

	runtime.ReadMemStats(&after)
	alive := runtime.NumGoroutine()
	stack := after.StackSys - before.StackSys
	other := (after.Sys - before.Sys) - stack // g structs, sudogs, run queues, ...

	fmt.Printf("goroutines alive: %d\n", alive)
	fmt.Printf("start time:       %v (%.0f ns per goroutine)\n",
		startTime.Round(time.Millisecond), float64(startTime.Nanoseconds())/float64(*n))
	fmt.Printf("stack memory:     %d MB (StackSys %d MB -> %d MB), %.1f KB per goroutine\n",
		stack>>20, before.StackSys>>20, after.StackSys>>20, float64(stack)/float64(*n)/1024)
	fmt.Printf("other memory:     %d MB, %.0f bytes per goroutine\n", other>>20, float64(other)/float64(*n))

	start = time.Now()
	close(stop) // wakes up every goroutine at once
	wg.Wait()
	fmt.Printf("after close:      %d goroutine(s), stopped in %v\n",
		runtime.NumGoroutine(), time.Since(start).Round(time.Millisecond))
}
