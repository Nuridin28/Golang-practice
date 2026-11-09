// Bonus: what does it print, and is it always the same?
//
//	go run ./bonus                 the program from the lesson
//	go run ./bonus -n 1000         1000 runs in this process: count the different orders
//	go run -race ./bonus -n 1000   the race detector shuffles the scheduler on purpose
package main

import (
	"flag"
	"fmt"
	"runtime"
	"sync"
)

// lesson is the program from the lesson.
func lesson() {
	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fmt.Print(i, " ")
		}()
	}
	wg.Wait()
	fmt.Println()
}

// order does the same, but collects the order instead of printing it.
// A buffered channel is safe to use from many goroutines.
func order() string {
	ch := make(chan int, 5)
	var wg sync.WaitGroup
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ch <- i
		}()
	}
	wg.Wait()
	close(ch)
	s := ""
	for i := range ch {
		s += fmt.Sprint(i, " ")
	}
	return s
}

func main() {
	n := flag.Int("n", 0, "run it n times and count the different orders")
	flag.Parse()
	runtime.GOMAXPROCS(1)

	if *n == 0 {
		lesson()
		return
	}
	seen := map[string]int{}
	for range *n {
		seen[order()]++
	}
	fmt.Printf("%d runs, %d different order(s):\n", *n, len(seen))
	for o, count := range seen {
		fmt.Printf("  %q x %d\n", o, count)
	}
}
