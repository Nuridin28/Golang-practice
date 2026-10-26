package main

import (
	"fmt"
	"strings"
)

func apply(nums []int, f func(int) int) []int {
	out := make([]int, len(nums))
	for i, n := range nums {
		out[i] = f(n)
	}
	return out
}

func double(n int) int { return n * 2 }

type Op func(a, b float64) float64

func divmod(a, b int) (q, r int) {
	q = a / b
	r = a % b
	return
}

func sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func counter() func() int {
	count := 0
	return func() int {
		count++
		return count
	}
}

func multiplier(k int) func(int) int {
	return func(n int) int { return n * k }
}

func main() {
	fmt.Println("--- functions as values")
	fmt.Println(apply([]int{1, 2, 3}, double))
	fmt.Println(apply([]int{1, 2, 3}, func(n int) int { return n * n }))
	var f func(string) string = strings.ToUpper
	fmt.Println(f("go"))

	var nilFunc func()
	fmt.Println("nilFunc == nil:", nilFunc == nil)
	func() {
		defer func() { fmt.Println("calling nil func:", recover()) }()
		nilFunc()
	}()

	fmt.Println("--- named function type")
	ops := map[string]Op{
		"+": func(a, b float64) float64 { return a + b },
		"*": func(a, b float64) float64 { return a * b },
	}
	fmt.Println(ops["+"](2, 3), ops["*"](2, 3))

	fmt.Println("--- multiple results, variadic")
	q, r := divmod(17, 5)
	fmt.Println(q, r)
	fmt.Println(sum(), sum(1, 2), sum([]int{4, 5, 6}...))

	fmt.Println("--- counter")
	c1 := counter()
	fmt.Println(c1(), c1(), c1())
	c2 := counter()
	fmt.Println(c2(), c1())

	fmt.Println("--- closures capture variables, not values")
	x := 10
	show := func() { fmt.Println("x =", x) }
	x = 20
	show()

	fmt.Println("--- multiplier")
	triple := multiplier(3)
	fmt.Println(triple(5))
}
