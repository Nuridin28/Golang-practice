package main

import (
	"fmt"
	"os"
	"slices"
	"strconv"
	"sync"
)

func deferArg() {
	for i := 0; i < 3; i++ {
		defer fmt.Print(i, " ")
	}
}

func deferClosure() {
	for i := 0; i < 3; i++ {
		defer func() { fmt.Print(i, " ") }()
	}
}

func ex08() {
	deferArg()
	fmt.Println()
	deferClosure()
	fmt.Println()
}

func ex09() {
	names := []string{"a", "b", "c"}
	var ptrs []*string
	for _, n := range names {
		ptrs = append(ptrs, &n)
	}
	for _, p := range ptrs {
		fmt.Print(*p, " ")
	}
	fmt.Println()
}

func ex10() {
	for i := 0; i < 5; i++ {
		fmt.Print(i, " ")
		i++
	}
	fmt.Println()
}

func ex11() {
	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		got []int
	)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			got = append(got, i)
			mu.Unlock()
		}()
	}
	wg.Wait()
	slices.Sort(got)
	fmt.Println(got)
}

func sumDeferred() (n int) {
	for i := 1; i <= 3; i++ {
		defer func() { n += i }()
	}
	return 10
}

func ex12() {
	fmt.Println(sumDeferred())
}

var exercises = map[int]func(){8: ex08, 9: ex09, 10: ex10, 11: ex11, 12: ex12}

func main() {
	n := 0
	if len(os.Args) == 2 {
		n, _ = strconv.Atoi(os.Args[1])
	}
	f, ok := exercises[n]
	if !ok {
		fmt.Println("usage: go run . N   (N = 8..12)")
		os.Exit(2)
	}
	f()
}
