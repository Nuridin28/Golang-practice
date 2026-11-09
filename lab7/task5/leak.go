package main

import (
	"context"
	"errors"
	"time"
)

func slowQuery() int {
	time.Sleep(50 * time.Millisecond) // pretend to ask a database
	return 42
}

// fetchLeaky: after a timeout nobody will ever read from ch,
// so the goroutine waits on `ch <-` forever.
func fetchLeaky(timeout time.Duration) (int, error) {
	ch := make(chan int) // unbuffered
	go func() {
		ch <- slowQuery() // blocks forever if fetchLeaky already returned
	}()
	select {
	case v := <-ch:
		return v, nil
	case <-time.After(timeout):
		return 0, errors.New("timeout")
	}
}

// fetchFixed: with a buffer of one the send always succeeds, even if nobody
// reads it, so the goroutine finishes as soon as slowQuery returns.
func fetchFixed(timeout time.Duration) (int, error) {
	ch := make(chan int, 1)
	go func() {
		ch <- slowQuery()
	}()
	select {
	case v := <-ch:
		return v, nil
	case <-time.After(timeout):
		return 0, errors.New("timeout")
	}
}

// startTickerLeaky starts a worker that nobody can stop.
func startTickerLeaky() {
	go func() {
		for {
			time.Sleep(10 * time.Millisecond) // "check something"
		}
	}()
}

// startTicker gives the caller a way to say "stop": cancel the context.
func startTicker(ctx context.Context) {
	go func() {
		t := time.NewTicker(10 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				// "check something"
			}
		}
	}()
}
