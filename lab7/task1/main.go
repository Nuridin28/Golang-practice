// Task 1: Concurrency vs parallelism.
//
//	go build -o /tmp/demo ./task1
//	/tmp/demo                    default GOMAXPROCS (= runtime.NumCPU())
//	GOMAXPROCS=1 /tmp/demo       only one goroutine runs Go code at a time
//	/tmp/demo -scale             cpuWork with 1 .. 2*NumCPU goroutines (Task 1.2)
//	/tmp/demo -v                 also print how long each goroutine worked
package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
	"time"
)

const N = 6_000_000

// isPrime is deliberately slow: trial division by every d with d*d <= n.
// A prime p costs about sqrt(p) divisions, so big numbers are harder to check.
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

// countPrimes counts the primes in [from, to).
func countPrimes(from, to int) int {
	count := 0
	for n := from; n < to; n++ {
		if isPrime(n) {
			count++
		}
	}
	return count
}

// cpuWork splits [0, N) into `workers` equal chunks, one goroutine per chunk.
// It also returns how long each goroutine worked.
func cpuWork(workers int) (int, time.Duration, []time.Duration) {
	start := time.Now()
	results := make([]int, workers)
	took := make([]time.Duration, workers)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		from, to := N*w/workers, N*(w+1)/workers // the last chunk ends exactly at N
		wg.Add(1)
		go func() {
			defer wg.Done()
			t := time.Now()
			results[w] = countPrimes(from, to) // each goroutine writes only its own element
			took[w] = time.Since(t)
		}()
	}
	wg.Wait()
	total := 0
	for _, r := range results {
		total += r
	}
	return total, time.Since(start), took
}

// ioWork starts `workers` goroutines that each "wait for a database" for 100 ms.
func ioWork(workers int) time.Duration {
	start := time.Now()
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(100 * time.Millisecond) // like waiting for a database
		}()
	}
	wg.Wait()
	return time.Since(start)
}

func main() {
	scale := flag.Bool("scale", false, "Task 1.2: cpuWork with 1 .. 2*NumCPU goroutines")
	verbose := flag.Bool("v", false, "print the time of every goroutine")
	flag.Parse()

	fmt.Printf("NumCPU=%d GOMAXPROCS=%d\n", runtime.NumCPU(), runtime.GOMAXPROCS(0))

	if *scale {
		scaling(*verbose)
		return
	}

	for _, w := range []int{1, 2, 4} {
		n, d, took := cpuWork(w)
		fmt.Printf("CPU work, %d goroutine(s): %d primes in %v\n", w, n, d.Round(time.Millisecond))
		if *verbose {
			printTook(took)
		}
	}
	for _, w := range []int{1, 10, 1000} {
		fmt.Printf("I/O work, %d goroutine(s): %v\n", w, ioWork(w).Round(time.Millisecond))
	}
}

// scaling runs cpuWork with more and more goroutines and prints the speed-up.
func scaling(verbose bool) {
	cpus := runtime.NumCPU()
	counts := []int{1, 2, 3, 4, 5, 6, 8}
	for _, c := range []int{cpus, cpus + cpus/2, 2 * cpus, 4 * cpus} {
		if c > counts[len(counts)-1] {
			counts = append(counts, c)
		}
	}
	var base time.Duration
	for _, w := range counts {
		n, d, took := cpuWork(w)
		if w == 1 {
			base = d
		}
		fmt.Printf("CPU work, %2d goroutine(s): %d primes in %-7v speed-up %.2fx\n",
			w, n, d.Round(time.Millisecond), float64(base)/float64(d))
		if verbose {
			printTook(took)
		}
	}
}

func printTook(took []time.Duration) {
	for i, t := range took {
		fmt.Printf("    goroutine %d: %v\n", i, t.Round(time.Millisecond))
	}
}
