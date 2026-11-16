package main

import (
	"runtime"
	"testing"
	"time"
)

func TestCollect(t *testing.T) {
	// workers w = 1..n each send w*10+0, w*10+1, w*10+2 → 30*w + 3.
	for _, n := range []int{1, 3, 10} {
		want := 0
		for w := 1; w <= n; w++ {
			want += 30*w + 3
		}
		if got := sumAll(collect(n)); got != want {
			t.Errorf("collect(%d): sum %d, want %d", n, got, want)
		}
	}
	if got := sumAll(collect(3)); got != 189 {
		t.Errorf("collect(3) = %d, want 189", got)
	}
}

func TestPipeline(t *testing.T) {
	if got := sum(square(generator(rangeN(10)...))); got != 385 { // 1+4+...+100
		t.Errorf("sum of squares = %d, want 385", got)
	}
	if got := sum(filterEven(square(generator(rangeN(10)...)))); got != 220 { // 4+16+36+64+100
		t.Errorf("sum of even squares = %d, want 220", got)
	}
	if got := sum(square(generator())); got != 0 {
		t.Errorf("empty pipeline = %d, want 0", got)
	}
}

// TestNoLeaks: when every stage closes its output, all goroutines end.
func TestNoLeaks(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 100; i++ {
		sum(filterEven(square(generator(rangeN(10)...))))
		sumAll(collect(5))
	}
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if n := runtime.NumGoroutine(); n > before {
		t.Errorf("%d goroutines leaked", n-before)
	}
}
