package task7

import "testing"

var sinkN int

// ids are 200 IDs: "id=<n>\n" for 1000..1199 is 8 bytes each, 1600 bytes.
var ids = func() []int {
	s := make([]int, 200)
	for i := range s {
		s[i] = 1000 + i
	}
	return s
}()

func TestRenderAgree(t *testing.T) {
	if a, b := RenderNew(ids), RenderPool(ids); a != 1600 || b != 1600 {
		t.Errorf("RenderNew = %d, RenderPool = %d, want 1600", a, b)
	}
}

func TestRenderAllocs(t *testing.T) {
	t.Logf("RenderNew:  %v allocs/op", testing.AllocsPerRun(1000, func() { sinkN = RenderNew(ids) }))
	t.Logf("RenderPool: %v allocs/op", testing.AllocsPerRun(1000, func() { sinkN = RenderPool(ids) }))
}

// TestNoReset shows why Reset is needed: the buffer keeps the old data.
// The pooled buffer still holds 1600 bytes from the earlier RenderPool calls
// (Reset happens after Get, not before Put), so every call grows it by 1600.
func TestNoReset(t *testing.T) {
	first := renderPoolNoReset(ids)
	second := renderPoolNoReset(ids)
	t.Logf("without Reset: first call %d bytes, second call %d bytes", first, second)
}

func BenchmarkRenderNew(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkN = RenderNew(ids)
	}
}

func BenchmarkRenderPool(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkN = RenderPool(ids)
	}
}

func BenchmarkRenderPoolParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		n := 0
		for pb.Next() {
			n += RenderPool(ids)
		}
		_ = n
	})
}

func BenchmarkRenderNewParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		n := 0
		for pb.Next() {
			n += RenderNew(ids)
		}
		_ = n
	})
}
