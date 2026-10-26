package main

import (
	"fmt"
	"time"
)

func counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func makeAccumulator() func(int) int {
	total := 0
	return func(n int) int {
		total += n
		return total
	}
}

func memoize(f func(int) int) func(int) int {
	cache := make(map[int]int)
	return func(n int) int {
		if v, ok := cache[n]; ok {
			return v
		}
		v := f(n)
		cache[n] = v
		return v
	}
}

func main() {
	fmt.Println("--- counter (participation task)")
	c1 := counter()
	fmt.Println(c1(), c1(), c1())
	c2 := counter()
	fmt.Println(c2(), c1())

	fmt.Println("--- makeAccumulator")
	acc := makeAccumulator()
	fmt.Println(acc(10), acc(5), acc(-3))
	other := makeAccumulator()
	fmt.Println(other(1), acc(0))

	fmt.Println("--- memoize")
	calls := make(map[int]int)
	slowSquare := func(n int) int {
		calls[n]++
		time.Sleep(100 * time.Millisecond)
		return n * n
	}
	fastSquare := memoize(slowSquare)
	inputs := []int{4, 4, 7, 4, 7, 9, 9}
	for _, n := range inputs {
		start := time.Now()
		v := fastSquare(n)
		fmt.Printf("square(%d) = %-3d %6.1fms\n", n, v, float64(time.Since(start).Microseconds())/1000)
	}
	total := 0
	for _, c := range calls {
		total += c
	}
	fmt.Printf("lookups: %d, slowSquare calls: %d, per input: %v\n", len(inputs), total, calls)

	fmt.Println("--- memoize a recursive function")
	fibCalls := 0
	var fib func(int) int
	fib = memoize(func(n int) int {
		fibCalls++
		if n < 2 {
			return n
		}
		return fib(n-1) + fib(n-2)
	})
	fmt.Printf("fib(90) = %d, calls: %d\n", fib(90), fibCalls)
}
