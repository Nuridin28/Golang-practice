package main

import (
	"errors"
	"runtime"
	"sort"
	"strings"
	"testing"
	"time"
)

func TestPoolFull(t *testing.T) {
	done := make(chan struct{})
	defer close(done)
	var got []string
	for r := range startWorkers(4, producer(words, done), done) {
		if r.Upper != strings.ToUpper(r.Word) {
			t.Errorf("bad result %+v", r)
		}
		got = append(got, r.Word)
	}
	sort.Strings(got)
	if strings.Join(got, ",") != strings.Join(words, ",") { // words is already sorted
		t.Errorf("got %v", got)
	}
}

func TestPoolCancelNoLeak(t *testing.T) {
	before := runtime.NumGoroutine()
	for i := 0; i < 100; i++ {
		done := make(chan struct{})
		results := startWorkers(4, producer(words, done), done)
		for k := 0; k < 3; k++ {
			<-results
		}
		close(done)
	}
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if n := runtime.NumGoroutine() - before; n > 0 {
		t.Errorf("%d goroutines leaked", n)
	}
}

func TestSubmit(t *testing.T) {
	q := make(chan string, 1)
	if err := Submit(q, "a", 10*time.Millisecond); err != nil {
		t.Fatalf("first submit: %v", err)
	}
	start := time.Now()
	if err := Submit(q, "b", 20*time.Millisecond); !errors.Is(err, ErrQueueFull) {
		t.Fatalf("second submit: %v, want ErrQueueFull", err)
	}
	if d := time.Since(start); d < 20*time.Millisecond {
		t.Errorf("Submit gave up after %v, before the timeout", d)
	}
	if v := <-q; v != "a" {
		t.Errorf("queue holds %q, want a", v)
	}
}
