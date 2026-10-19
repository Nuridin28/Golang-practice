// Task 5: Generic toolbox.
package main

import (
	"cmp"
	"fmt"
)

// Without the ~ (int | int64 | float64) every call with Score fails:
// Score does not satisfy Number (possibly missing ~ for int in Number)
type Number interface {
	~int | ~int64 | ~float64
}

type Score int

func Sum[T Number](nums []T) T {
	var s T
	for _, n := range nums {
		s += n
	}
	return s
}

// Average returns float64, not T: the average of ints 1 and 2 is 1.5.
func Average[T Number](nums []T) (float64, bool) {
	if len(nums) == 0 {
		return 0, false
	}
	return float64(Sum(nums)) / float64(len(nums)), true
}

func MaxOf[T cmp.Ordered](items []T) (T, bool) {
	var zero T
	if len(items) == 0 {
		return zero, false
	}
	m := items[0]
	for _, v := range items[1:] {
		if v > m {
			m = v
		}
	}
	return m, true
}

func IndexOf[T comparable](items []T, target T) int {
	for i, v := range items {
		if v == target {
			return i
		}
	}
	return -1
}

func Filter[T any](items []T, keep func(T) bool) []T {
	var out []T
	for _, v := range items {
		if keep(v) {
			out = append(out, v)
		}
	}
	return out
}

func main() {
	scores := []Score{90, 60, 75, 88, 74}

	fmt.Println("Sum ints:      ", Sum([]int{1, 2, 3}))
	fmt.Println("Sum floats:    ", Sum([]float64{1.5, 2.5}))
	fmt.Println("Sum scores:    ", Sum(scores))
	fmt.Printf("Sum type:       %T\n", Sum(scores))

	fmt.Println(Average(scores))
	fmt.Println(Average([]int{1, 2}))
	fmt.Println(Average([]Score{}))

	fmt.Println(MaxOf([]int{4, 9, 2}))
	fmt.Println(MaxOf([]string{"go", "rust", "c"}))
	fmt.Println(MaxOf([]int{}))
	fmt.Println(MaxOf(scores))

	fmt.Println(IndexOf(scores, 75))
	fmt.Println(IndexOf(scores, 100))
	fmt.Println(IndexOf([]string{"go", "rust"}, "rust"))

	passed := Filter(scores, func(s Score) bool { return s >= 75 })
	fmt.Println("passed:", passed)
}
