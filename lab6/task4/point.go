// Task 4: Benchmark — value vs pointer, with vs without capacity.
// Run: go test -run='^$' -bench=. -benchmem ./task4
//
//	go build -o /dev/null -gcflags='-m -l' ./task4
package task4

type Point struct{ X, Y int }

// Both constructors are marked noinline: if NewPointPointer were inlined
// into the benchmark loop, escape analysis would see that the pointer never
// leaves the loop and put the Point on the stack. Then the benchmark would
// measure nothing.

//go:noinline
func NewPointValue(x, y int) Point { return Point{x, y} }

//go:noinline
func NewPointPointer(x, y int) *Point { return &Point{x, y} }

// The Collect functions are noinline too, so that the benchmark and
// testing.AllocsPerRun measure the same code. Inlined into the benchmark
// loop, CollectNoCap loses the 32-byte stack buffer for its first append
// (Go 1.25+) and makes 12 allocations instead of 9.

//go:noinline
func CollectNoCap(n int) []int {
	var s []int
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}

//go:noinline
func CollectWithCap(n int) []int {
	s := make([]int, 0, n)
	for i := 0; i < n; i++ {
		s = append(s, i)
	}
	return s
}
