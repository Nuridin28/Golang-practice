// Package task6 compares a concrete, a generic and an any-based sum (Task 6).
package task6

type Number interface {
	~int | ~int64 | ~float64
}

func SumInts(nums []int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

func Sum[T Number](nums []T) T {
	var total T
	for _, n := range nums {
		total += n
	}
	return total
}

func SumAny(nums []any) int {
	total := 0
	for _, n := range nums {
		total += n.(int) // type assertion on every element
	}
	return total
}

// ToAny copies every int into a []any. Each int that is put into an
// interface is boxed: it may need its own heap allocation.
func ToAny(nums []int) []any {
	out := make([]any, len(nums))
	for i, n := range nums {
		out[i] = n
	}
	return out
}
