package main

import (
	"fmt"
	"time"
)

// The Part 4 experiment. The producer makes 2000 jobs in bursts of 50 every
// 10 ms (≈ 5 jobs/ms). The consumer handles 1 job/ms.
//
//	busy    = how long the producer needed to hand over all its jobs
//	queue   = max jobs waiting in the channel
//	wait    = the longest time one job spent in the queue
const (
	totalJobs = 2000
	burst     = 50
	every     = 10 * time.Millisecond
	work      = time.Millisecond
)

type Job struct {
	ID      int
	Created time.Time
}

type stats struct {
	busy             time.Duration
	maxQueue         int
	handled, dropped int
	maxWait          time.Duration
}

// strategies: how the producer puts one job into the queue.
// It returns false if the job was dropped.
var strategies = []struct {
	name string
	size int
	put  func(jobs chan Job, j Job) bool
}{
	{"blocking", 10, func(jobs chan Job, j Job) bool { jobs <- j; return true }},
	{"huge", totalJobs, func(jobs chan Job, j Job) bool { jobs <- j; return true }},
	{"dropping", 10, func(jobs chan Job, j Job) bool {
		select {
		case jobs <- j:
			return true
		default:
			return false
		}
	}},
}

func backpressure() {
	fmt.Println("# busy = how long the producer needed · queue = max jobs waiting · wait = longest wait of one job")
	for _, s := range strategies {
		st := experiment(s.size, s.put)
		fmt.Printf("%-9s busy %-7v queue %-5d handled %-5d dropped %-5d wait %v\n",
			s.name, st.busy.Round(10*time.Millisecond), st.maxQueue, st.handled, st.dropped, st.maxWait.Round(10*time.Millisecond))
	}
}

func experiment(size int, put func(chan Job, Job) bool) stats {
	var st stats
	jobs := make(chan Job, size)

	consumerDone := make(chan struct{})
	go func() { // the only writer of handled and maxWait; main reads them after <-consumerDone
		defer close(consumerDone)
		for j := range jobs {
			if w := time.Since(j.Created); w > st.maxWait {
				st.maxWait = w
			}
			time.Sleep(work)
			st.handled++
		}
	}()

	start := time.Now()
	next := start
	for id := 0; id < totalJobs; {
		for k := 0; k < burst && id < totalJobs; k++ {
			if !put(jobs, Job{ID: id, Created: time.Now()}) {
				st.dropped++
			}
			if l := len(jobs); l > st.maxQueue { // len is only a monitoring number here, never a decision
				st.maxQueue = l
			}
			id++
		}
		next = next.Add(every)
		if d := time.Until(next); d > 0 {
			time.Sleep(d) // wait for the next burst (if we're already late, don't wait)
		}
	}
	st.busy = time.Since(start)
	close(jobs)
	<-consumerDone
	return st
}
