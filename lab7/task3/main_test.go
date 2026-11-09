package main

import (
	"sync"
	"testing"
)

var sink int

//go:noinline
func work(i int) int { return i * 2 }

// BenchmarkGoroutineStartStop: start a goroutine and wait until it has finished.
func BenchmarkGoroutineStartStop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			sink = work(i)
		}()
		wg.Wait()
	}
}

// BenchmarkFuncCall: the same work as a plain function call.
func BenchmarkFuncCall(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = work(i)
	}
}

// BenchmarkGoroutineStartStop100: start 100 goroutines and wait for all of them.
// Here the starts overlap and run on several cores, so one start is cheaper.
func BenchmarkGoroutineStartStop100(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		for j := 0; j < 100; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_ = work(j)
			}()
		}
		wg.Wait()
	}
}
