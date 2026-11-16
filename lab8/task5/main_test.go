package main

import (
	"errors"
	"runtime"
	"slices"
	"testing"
	"time"
)

func TestMerge(t *testing.T) {
	var got []int
	for v := range merge(source(1, 3, time.Millisecond), source(100, 5, 0)) {
		got = append(got, v)
	}
	slices.Sort(got)
	want := []int{1, 2, 3, 100, 101, 102, 103, 104}
	if !slices.Equal(got, want) {
		t.Errorf("merge = %v, want %v", got, want)
	}
}

func TestMergeEmpty(t *testing.T) {
	for v := range merge(source(0, 0, 0), source(0, 0, 0)) {
		t.Errorf("unexpected value %d", v)
	}
}

func TestFetchWithTimeout(t *testing.T) {
	slow := func() int { time.Sleep(30 * time.Millisecond); return 42 }
	if v, err := fetchWithTimeout(slow, time.Second); v != 42 || err != nil {
		t.Errorf("got %d, %v; want 42, nil", v, err)
	}
	before := runtime.NumGoroutine()
	for i := 0; i < 100; i++ {
		if _, err := fetchWithTimeout(slow, time.Millisecond); !errors.Is(err, ErrTimeout) {
			t.Fatalf("err = %v, want ErrTimeout", err)
		}
	}
	time.Sleep(100 * time.Millisecond) // every slow() has returned
	if n := runtime.NumGoroutine() - before; n > 0 {
		t.Errorf("%d goroutines leaked", n)
	}
}

func TestPreferA(t *testing.T) {
	c := preferA(1000)
	if c["a"] != 500 || c["b"] != 500 {
		t.Errorf("preferA = %v, want all 500 a's first, then 500 b's", c)
	}
}
