package task4

import "testing"

var (
	sinkPoint Point
	sinkPtr   *Point
	sinkInts  []int
)

func TestAllocs(t *testing.T) {
	cases := []struct {
		name string
		f    func()
	}{
		{"NewPointValue", func() { sinkPoint = NewPointValue(1, 2) }},
		{"NewPointPointer", func() { sinkPtr = NewPointPointer(1, 2) }},
		{"CollectNoCap(1000)", func() { sinkInts = CollectNoCap(1000) }},
		{"CollectWithCap(1000)", func() { sinkInts = CollectWithCap(1000) }},
	}
	for _, c := range cases {
		t.Logf("%-22s %v allocs/op", c.name, testing.AllocsPerRun(1000, c.f))
	}
}

// TestCapGrowth prints every capacity CollectNoCap goes through.
func TestCapGrowth(t *testing.T) {
	var s []int
	prev, total := -1, 0
	for i := 0; i < 1000; i++ {
		s = append(s, i)
		if cap(s) != prev {
			prev = cap(s)
			total += cap(s) * 8
			t.Logf("len %4d -> cap %4d (%5d bytes)", len(s), cap(s), cap(s)*8)
		}
	}
	t.Logf("sum of all arrays: %d bytes", total)
}

func BenchmarkNewPointValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkPoint = NewPointValue(i, i)
	}
}

func BenchmarkNewPointPointer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkPtr = NewPointPointer(i, i)
	}
}

func BenchmarkCollectNoCap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkInts = CollectNoCap(1000)
	}
}

func BenchmarkCollectWithCap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkInts = CollectWithCap(1000)
	}
}
