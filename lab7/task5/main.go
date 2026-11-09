// Task 5: Find and fix the leaks.
//
//	go run ./task5                  count goroutines after every step
//	go run ./task5 -profile 1       + the goroutine profile (debug=1): where they are stuck
//	go run ./task5 -profile 2       + debug=2: every goroutine with its state
//	go run ./task5 -leakprofile     + the goroutineleak profile (Go 1.26 experiment, Go 1.27 GA)
//
//	go test -v ./task5                          fetchFixed and startTicker don't leak
//	SHOW_LEAK=1 go test -run Leaky -v ./task5   the same check fails for fetchLeaky
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"
)

func main() {
	profile := flag.Int("profile", 0, "print the goroutine profile with this debug level (1 or 2)")
	leakProfile := flag.Bool("leakprofile", false, "print the goroutineleak profile")
	flag.Parse()

	fmt.Println("start:", runtime.NumGoroutine())

	for i := 1; i <= 100; i++ {
		fetchLeaky(10 * time.Millisecond) // slowQuery takes 50 ms, so every call times out
		if i%25 == 0 {
			fmt.Printf("after %d fetchLeaky: %d\n", i, runtime.NumGoroutine())
		}
	}

	for i := 1; i <= 100; i++ {
		fetchFixed(10 * time.Millisecond)
	}
	fmt.Printf("after 100 fetchFixed: %d (the last ones are still in slowQuery)\n", runtime.NumGoroutine())
	time.Sleep(100 * time.Millisecond)
	fmt.Printf("100 ms later:         %d (they all sent into the buffer and returned)\n", runtime.NumGoroutine())

	for i := 0; i < 10; i++ {
		startTickerLeaky()
	}
	fmt.Println("after 10 leaky tickers:", runtime.NumGoroutine())

	ctx, cancel := context.WithCancel(context.Background())
	for i := 0; i < 10; i++ {
		startTicker(ctx)
	}
	fmt.Println("after 10 stoppable tickers:", runtime.NumGoroutine())

	if *profile > 0 {
		fmt.Printf("\n===== goroutine profile, debug=%d =====\n", *profile)
		pprof.Lookup("goroutine").WriteTo(os.Stdout, *profile)
		fmt.Println("===== end of profile =====")
	}
	if *leakProfile {
		// Writing this profile first runs a GC that marks goroutines blocked on
		// channels/locks that nothing reachable can ever wake up.
		fmt.Println("\n===== goroutineleak profile, debug=1 =====")
		pprof.Lookup("goroutineleak").WriteTo(os.Stdout, 1)
		fmt.Println("===== end of profile =====")
	}

	cancel()
	time.Sleep(50 * time.Millisecond) // let the stoppable tickers see ctx.Done() and return
	fmt.Println("after cancel():", runtime.NumGoroutine())
}
