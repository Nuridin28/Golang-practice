// Task 6: A leaking "server".
//
//	go run ./task6 leaky       unbuffered channel: every timeout leaks a goroutine + 64 KB
//	go run ./task6 buffered    fix (a): make(chan *Report, 1)
//	go run ./task6 ctx         fix (b): buffered channel + context.WithTimeout passed to buildReport
//
// Every round starts 1000 handle calls at once, each with a 5 ms timeout.
// buildReport needs 20 ms, so every call times out.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type Report struct {
	Data [64 << 10]byte // 64 KB
}

var built, abandoned atomic.Int64 // reports built completely / work stopped early

// newReport allocates and fills the 64 KB report.
func newReport() *Report {
	r := new(Report)
	for i := range r.Data {
		r.Data[i] = byte(i)
	}
	built.Add(1)
	return r
}

// buildReport "queries a database" for 20 ms and then builds a 64 KB report.
// It can't be stopped: it always does all the work.
func buildReport() *Report {
	time.Sleep(20 * time.Millisecond)
	return newReport()
}

// buildReportCtx does the same, but gives up as soon as ctx is done:
// no more waiting, and the 64 KB report is never built.
// (A real database driver takes ctx too: db.QueryContext(ctx, ...).)
func buildReportCtx(ctx context.Context) (*Report, error) {
	t := time.NewTimer(20 * time.Millisecond)
	defer t.Stop()
	select {
	case <-t.C:
	case <-ctx.Done():
		abandoned.Add(1)
		return nil, ctx.Err()
	}
	return newReport(), nil
}

// handleLeaky: after a timeout nobody reads ch, the goroutine is stuck on
// `ch <-` forever, and its stack keeps the 64 KB report reachable.
func handleLeaky(timeout time.Duration) error {
	ch := make(chan *Report)
	go func() { ch <- buildReport() }()
	select {
	case r := <-ch:
		_ = r
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

// handleBuffered: fix (a). The send never blocks, so the goroutine ends
// after 20 ms. But it still does all the work for a result nobody wants.
func handleBuffered(timeout time.Duration) error {
	ch := make(chan *Report, 1)
	go func() { ch <- buildReport() }()
	select {
	case r := <-ch:
		_ = r
		return nil
	case <-time.After(timeout):
		return errors.New("timeout")
	}
}

// handleCtx: fix (b). One context controls both the waiting here and the work
// in the goroutine. When handle returns, cancel() stops buildReportCtx at once.
func handleCtx(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	type result struct {
		r   *Report
		err error
	}
	ch := make(chan result, 1) // still buffered: the goroutine may finish after we return
	go func() {
		r, err := buildReportCtx(ctx)
		ch <- result{r, err}
	}()
	select {
	case res := <-ch:
		return res.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./task6 leaky|buffered|ctx")
		os.Exit(2)
	}
	var handle func(time.Duration) error
	switch os.Args[1] {
	case "leaky":
		handle = handleLeaky
	case "buffered":
		handle = handleBuffered
	case "ctx":
		handle = handleCtx
	default:
		fmt.Println("unknown mode", os.Args[1])
		os.Exit(2)
	}

	fmt.Printf("mode=%s: 3 rounds x 1000 handle calls, timeout 5 ms, report 64 KB in 20 ms\n", os.Args[1])
	printStats("start", 0, nil)
	for round := 1; round <= 3; round++ {
		var before runtime.MemStats
		runtime.ReadMemStats(&before)

		var wg sync.WaitGroup
		var timeouts atomic.Int64
		for i := 0; i < 1000; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				if handle(5*time.Millisecond) != nil {
					timeouts.Add(1)
				}
			}()
		}
		wg.Wait() // all 1000 handle calls have returned
		inFlight := runtime.NumGoroutine()

		time.Sleep(100 * time.Millisecond) // everything that can finish has finished now
		printStats(fmt.Sprintf("after round %d", round), inFlight, &before)
	}
	fmt.Printf("reports built: %d, builds stopped early: %d\n", built.Load(), abandoned.Load())
}

// printStats prints the goroutines and the live heap (after a GC).
// inFlight is the number of goroutines right after the 1000 handle calls returned.
func printStats(label string, inFlight int, before *runtime.MemStats) {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("%-14s goroutines=%-5d heap=%4d MB", label, runtime.NumGoroutine(), m.HeapAlloc>>20)
	if before != nil {
		fmt.Printf("   right after handle returned: %4d goroutines   allocated in round: %3d MB",
			inFlight, (m.TotalAlloc-before.TotalAlloc)>>20)
	}
	fmt.Println()
}
