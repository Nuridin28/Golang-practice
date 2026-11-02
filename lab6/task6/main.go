// Task 6: Watch the GC. ~22 MB of live data and ~2.7 GB of short-lived garbage.
//
//	go build -o /tmp/gcdemo ./task6
//	GOGC=50 /tmp/gcdemo     (100, 200, 400, off)
//	GOGC=off GOMEMLIMIT=100MiB /tmp/gcdemo
//	GODEBUG=gctrace=1 /tmp/gcdemo
//
// Or all runs at once: ./task6/run.sh
package main

import (
	"fmt"
	"os"
	"runtime"
	"time"
)

type Record struct {
	ID   int
	Data [256]byte
}

var live []*Record // stays reachable: the live heap

// lastSeen makes every Record escape to the heap. Without it consume only
// reads r, escape analysis keeps &Record{} on the stack, and the program
// makes no garbage at all (go build -gcflags=-m: "&Record{...} does not escape").
var lastSeen *Record

//go:noinline
func consume(r *Record) int {
	lastSeen = r
	return r.ID + int(r.Data[0])
}

func main() {
	start := time.Now()
	for i := 0; i < 80_000; i++ {
		live = append(live, &Record{ID: i})
	}
	sum := 0
	for i := 0; i < 10_000_000; i++ {
		r := &Record{ID: i} // garbage after one use
		sum += consume(r)
	}
	elapsed := time.Since(start)

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("GOGC=%-4s GOMEMLIMIT=%-7s cycles=%-4d heapSys=%4d MB  pauses=%5.1f ms  time=%.2fs  (sum %d)\n",
		env("GOGC", "100"), env("GOMEMLIMIT", "-"),
		m.NumGC, m.HeapSys>>20, float64(m.PauseTotalNs)/1e6, elapsed.Seconds(), sum%10)
}

func env(name, def string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return def
}
