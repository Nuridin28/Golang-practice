package task6

import "testing"

const size = 1000

// sink keeps the compiler from throwing away the results.
var sink int

// numbers returns 0, 1, ..., n-1.
func numbers(n int) []int {
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i
	}
	return nums
}

func TestSumsAgree(t *testing.T) {
	nums := numbers(size)
	want := size * (size - 1) / 2
	if got := SumInts(nums); got != want {
		t.Errorf("SumInts = %d, want %d", got, want)
	}
	if got := Sum(nums); got != want {
		t.Errorf("Sum = %d, want %d", got, want)
	}
	if got := SumAny(ToAny(nums)); got != want {
		t.Errorf("SumAny = %d, want %d", got, want)
	}
}

func BenchmarkSumInts(b *testing.B) {
	nums := numbers(size)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = SumInts(nums)
	}
}

func BenchmarkSumGeneric(b *testing.B) {
	nums := numbers(size)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = Sum(nums)
	}
}

func BenchmarkSumAny(b *testing.B) {
	anys := ToAny(numbers(size))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = SumAny(anys)
	}
}

func BenchmarkSumAnyWithBoxing(b *testing.B) {
	nums := numbers(size)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sink = SumAny(ToAny(nums))
	}
}
