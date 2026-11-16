package main

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Word, Upper string
}

// producer owns jobs: it creates it, it is the only writer, it closes it.
// It stops early when done is closed.
func producer(words []string, done <-chan struct{}) <-chan string {
	jobs := make(chan string)
	go func() {
		defer close(jobs)
		for _, w := range words {
			select {
			case jobs <- w:
			case <-done: // the consumer gave up: stop producing
				return
			}
		}
	}()
	return jobs
}

// startWorkers starts n workers that turn jobs into Results. results has many
// writers, so none of them closes it: a closer goroutine waits for all workers.
func startWorkers(n int, jobs <-chan string, done <-chan struct{}) <-chan Result {
	results := make(chan Result)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for w := range jobs { // ends when the producer closes jobs
				r := Result{Word: w, Upper: strings.ToUpper(w)}
				select {
				case results <- r:
				case <-done: // nobody reads results any more
					return
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(results)
	}()
	return results
}

// ---------------------------------------------------------------- the leaky versions (no <-done cases)

func producerLeaky(words []string) <-chan string {
	jobs := make(chan string)
	go func() {
		defer close(jobs)
		for _, w := range words {
			jobs <- w // blocks forever once the workers are stuck
		}
	}()
	return jobs
}

func startWorkersLeaky(n int, jobs <-chan string) <-chan Result {
	results := make(chan Result)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for w := range jobs {
				results <- Result{Word: w, Upper: strings.ToUpper(w)} // blocks forever once main stops reading
			}
		}()
	}
	go func() {
		wg.Wait() // never returns: the workers never finish
		close(results)
	}()
	return results
}

// ---------------------------------------------------------------- Submit

var ErrQueueFull = errors.New("queue is full")

// Submit waits at most `timeout` for a free slot in queue. Block a little,
// then give up and tell the caller (an HTTP server would answer 503).
func Submit(queue chan<- string, job string, timeout time.Duration) error {
	t := time.NewTimer(timeout)
	defer t.Stop()
	select {
	case queue <- job:
		return nil
	case <-t.C:
		return ErrQueueFull
	}
}
