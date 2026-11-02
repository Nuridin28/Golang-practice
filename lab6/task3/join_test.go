package task3

import "testing"

var sinkS string

// ids are 10 IDs. Values > 255 on purpose: smaller ints are boxed into an
// interface without an allocation (runtime.staticuint64s, Lab 4).
var ids = []int{1001, 1002, 1003, 1004, 1005, 1006, 1007, 1008, 1009, 1010}

const want = "1001,1002,1003,1004,1005,1006,1007,1008,1009,1010"

func TestJoinIDsAgree(t *testing.T) {
	if got := JoinIDsSlow(ids); got != want {
		t.Errorf("JoinIDsSlow = %q, want %q", got, want)
	}
	if got := JoinIDsFast(ids); got != want {
		t.Errorf("JoinIDsFast = %q, want %q", got, want)
	}
	if got := JoinIDsFast(nil); got != "" {
		t.Errorf("JoinIDsFast(nil) = %q, want empty", got)
	}
}

func TestJoinIDsSlowAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		sinkS = JoinIDsSlow(ids)
	})
	t.Logf("JoinIDsSlow: %v allocs for %d IDs", allocs, len(ids))
}

func TestJoinIDsFastAllocs(t *testing.T) {
	allocs := testing.AllocsPerRun(1000, func() {
		sinkS = JoinIDsFast(ids)
	})
	t.Logf("JoinIDsFast: %v allocs for %d IDs", allocs, len(ids))
	if allocs > 1 {
		t.Errorf("JoinIDsFast: %v allocs, want at most 1", allocs)
	}
}

func BenchmarkJoinIDsSlow(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkS = JoinIDsSlow(ids)
	}
}

func BenchmarkJoinIDsFast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkS = JoinIDsFast(ids)
	}
}
