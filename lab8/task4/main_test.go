package main

import (
	"runtime"
	"slices"
	"testing"
	"time"

	"lab8/task4/tree"
)

func TestWalk(t *testing.T) {
	for k := 1; k <= 5; k++ {
		ch := make(chan int)
		go Walk(tree.New(k), ch)
		var got []int
		for v := range ch {
			got = append(got, v)
		}
		want := []int{k, 2 * k, 3 * k, 4 * k, 5 * k, 6 * k, 7 * k, 8 * k, 9 * k, 10 * k}
		if !slices.Equal(got, want) {
			t.Errorf("Walk(New(%d)) = %v, want %v", k, got, want)
		}
	}
	ch := make(chan int)
	go Walk(nil, ch)
	if _, ok := <-ch; ok {
		t.Error("Walk(nil) sent a value")
	}
}

func TestSame(t *testing.T) {
	cases := []struct {
		a, b *tree.Tree
		want bool
	}{
		{tree.New(1), tree.New(1), true},
		{tree.New(3), tree.New(3), true},
		{tree.New(1), tree.New(2), false},
		{tree.FromValues(1, 2, 3), tree.FromValues(3, 2, 1), true},
		{tree.FromValues(1, 2, 3), tree.FromValues(1, 2), false},    // b is a prefix of a
		{tree.FromValues(1, 2), tree.FromValues(1, 2, 3), false},    // a is a prefix of b
		{tree.FromValues(1, 2, 3), tree.FromValues(1, 2, 4), false}, // differs at the end
		{nil, nil, true},
	}
	for i, c := range cases {
		if got := Same(c.a, c.b); got != c.want {
			t.Errorf("case %d: Same = %v, want %v", i, got, c.want)
		}
		if got := SameLeaky(c.a, c.b); got != c.want {
			t.Errorf("case %d: SameLeaky = %v, want %v", i, got, c.want)
		}
	}
}

func goroutinesAfter(f func()) int {
	before := runtime.NumGoroutine()
	f()
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	return runtime.NumGoroutine() - before
}

func TestSameNoLeak(t *testing.T) {
	leaked := goroutinesAfter(func() {
		for i := 0; i < 100; i++ {
			Same(tree.New(1), tree.New(2))
		}
	})
	if leaked > 0 {
		t.Errorf("Same leaked %d goroutines", leaked)
	}
}

// TestSameLeakyLeaks documents the bug: SameLeaky leaves 2 walkers per call.
func TestSameLeakyLeaks(t *testing.T) {
	leaked := goroutinesAfter(func() {
		for i := 0; i < 100; i++ {
			SameLeaky(tree.New(1), tree.New(2))
		}
	})
	if leaked != 200 {
		t.Errorf("SameLeaky: %d goroutines left, expected 200 (2 per call)", leaked)
	}
}
