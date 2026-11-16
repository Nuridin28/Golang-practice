// Task 6: A worker pool with backpressure and cancellation.
//
//	go run ./task6 full          run the pool to the end with range
//	go run ./task6 take3         take 3 results, close(done): NumGoroutine must go back to 1
//	go run ./task6 leaky         the same without the <-done cases: count the leaked goroutines
//	go run ./task6 submit        Submit: one success and one ErrQueueFull
//	go run ./task6 backpressure  the Part 4 experiment: block / huge buffer / drop
//
// Tests: go test -race -v ./task6
package main

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
)

var words = []string{
	"alpha", "bravo", "charlie", "delta", "echo", "foxtrot", "golf", "hotel",
	"india", "juliett", "kilo", "lima", "mike", "november", "oscar", "papa",
	"quebec", "romeo", "sierra", "tango", "uniform", "victor", "whiskey", "xray",
	"yankee", "zulu",
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./task6 full|take3|leaky|submit|backpressure")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "full":
		full()
	case "take3":
		take3()
	case "leaky":
		leaky()
	case "submit":
		submitDemo()
	case "backpressure":
		backpressure()
	default:
		fmt.Println("unknown mode", os.Args[1])
		os.Exit(2)
	}
}

func full() {
	done := make(chan struct{})
	defer close(done)
	n := 0
	for r := range startWorkers(4, producer(words, done), done) {
		fmt.Printf("%s→%s ", r.Word, r.Upper)
		n++
	}
	fmt.Printf("\n%d results; goroutines now: %d\n", n, runtime.NumGoroutine())
}

// settle waits (at most 1 s) until the number of goroutines stops above want.
// The stopped goroutines need a moment to be scheduled and return.
func settle(want int) int {
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > want && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	return runtime.NumGoroutine()
}

func take3() {
	done := make(chan struct{})
	results := startWorkers(4, producer(words, done), done)
	fmt.Println("goroutines while running:", runtime.NumGoroutine(), "(main + producer + 4 workers + closer)")
	for i := 0; i < 3; i++ {
		r := <-results
		fmt.Println("result:", r.Word, "→", r.Upper)
	}
	close(done) // stop everybody: producer, workers, and through them the closer
	fmt.Println("goroutines after close(done):", settle(1))
}

func leaky() {
	results := startWorkersLeaky(4, producerLeaky(words))
	for i := 0; i < 3; i++ {
		r := <-results
		fmt.Println("result:", r.Word, "→", r.Upper)
	}
	// main walks away. Nobody can tell the others to stop.
	n := settle(1)
	fmt.Printf("goroutines after taking 3 results: %d → %d leaked\n", n, n-1)
	buf := make([]byte, 1<<16)
	buf = buf[:runtime.Stack(buf, true)]
	fmt.Println("where they are stuck:")
	for _, line := range strings.Split(string(buf), "\n") {
		if strings.HasPrefix(line, "goroutine ") {
			fmt.Println("  ", line)
		}
	}
}

func submitDemo() {
	queue := make(chan string, 1) // one slot, and nobody is consuming
	err := Submit(queue, "job-1", 50*time.Millisecond)
	fmt.Println("submit job-1:", errOK(err))
	t := time.Now()
	err = Submit(queue, "job-2", 50*time.Millisecond)
	fmt.Printf("submit job-2: %s (after %v)\n", errOK(err), time.Since(t).Round(time.Millisecond))
	if errors.Is(err, ErrQueueFull) {
		fmt.Println("→ an HTTP handler would answer 503 Service Unavailable + Retry-After")
	}

	// And with a consumer that frees the slot in time:
	go func() { time.Sleep(20 * time.Millisecond); <-queue }()
	t = time.Now()
	err = Submit(queue, "job-3", 50*time.Millisecond)
	fmt.Printf("submit job-3 (a consumer frees a slot after 20ms): %s (after %v)\n", errOK(err), time.Since(t).Round(time.Millisecond))
}

func errOK(err error) string {
	if err == nil {
		return "ok"
	}
	return "error: " + err.Error()
}
