// Task 3: Closing and ownership.
//
//	go run ./task3 wrong        3.1: every producer closes the shared channel (panics)
//	go run ./task3 collect      3.1: one closer goroutine waits for all writers, then closes
//	go run -race ./task3 collect
//	go run ./task3 closeinline  question: wg.Wait(); close(results) inside collect, before return (deadlock)
//	go run ./task3 pipeline     3.2: generator(1..10) -> square -> sum
//	go run ./task3 pipeline4    3.3: generator -> square -> filterEven -> sum
//
// Tests: go test -race -v ./task3
package main

import (
	"fmt"
	"os"
	"sync"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: go run ./task3 wrong|collect|closeinline|pipeline|pipeline4")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "wrong":
		wrong()
	case "collect":
		fmt.Println("sum:", sumAll(collect(3)))
	case "closeinline":
		fmt.Println("sum:", sumAll(collectCloseInline(3)))
	case "pipeline":
		fmt.Println("sum of squares 1..10:", sum(square(generator(rangeN(10)...))))
	case "pipeline4":
		fmt.Println("sum of even squares 1..10:", sum(filterEven(square(generator(rangeN(10)...)))))
	default:
		fmt.Println("unknown mode", os.Args[1])
		os.Exit(2)
	}
}

// wrong: every producer closes the shared channel when IT is done.
func wrong() {
	results := make(chan int)
	var wg sync.WaitGroup
	for w := 1; w <= 3; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 3; i++ {
				results <- w*10 + i
			}
			close(results) // "I'm done" — but the others are not!
		}()
	}
	sum := 0
	for v := range results {
		sum += v
	}
	fmt.Println("sum:", sum)
	wg.Wait()
}

// collect: many writers, none of them closes. One extra goroutine waits for
// all of them and closes exactly once.
func collect(workers int) <-chan int {
	results := make(chan int)
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 3; i++ {
				results <- w*10 + i
			}
		}()
	}
	go func() {
		wg.Wait()      // all writers are finished...
		close(results) // ...so now it's safe to close
	}()
	return results
}

// collectCloseInline: the question from the tutorial. wg.Wait() runs before
// anybody can receive: the writers block on the unbuffered `results <-`,
// wg.Wait() never returns and `return results` is never reached.
func collectCloseInline(workers int) <-chan int {
	results := make(chan int)
	var wg sync.WaitGroup
	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 3; i++ {
				results <- w*10 + i
			}
		}()
	}
	wg.Wait()
	close(results)
	return results
}

func sumAll(in <-chan int) int {
	s := 0
	for v := range in {
		s += v
	}
	return s
}

func rangeN(n int) []int {
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i + 1
	}
	return nums
}
