package task5

import "testing"

// sink keeps the compiler from throwing away the results.
var sink float64

func BenchmarkFirstValue(b *testing.B) {
	r := Report{Title: "big"}
	for i := 0; i < b.N; i++ {
		sink = r.FirstValue()
	}
}

func BenchmarkFirstPointer(b *testing.B) {
	r := &Report{Title: "big"}
	for i := 0; i < b.N; i++ {
		sink = r.FirstPointer()
	}
}

func BenchmarkPointValue(b *testing.B) {
	p := Point{1, 2}
	for i := 0; i < b.N; i++ {
		sink = p.XValue()
	}
}

func BenchmarkPointPointer(b *testing.B) {
	p := &Point{1, 2}
	for i := 0; i < b.N; i++ {
		sink = p.XPointer()
	}
}
