package task7

import (
	"fmt"
	"testing"
	"time"
)

const N = 200_000

var sink int

func TestCountSeq(t *testing.T) {
	// π(n): the number of primes below n.
	for n, want := range map[int]int{0: 0, 1: 0, 2: 0, 3: 1, 10: 4, 100: 25, 1000: 168, N: 17984} {
		if got := CountSeq(n); got != want {
			t.Errorf("CountSeq(%d) = %d, want %d", n, got, want)
		}
	}
}

func TestCountParEqualsSeq(t *testing.T) {
	for _, n := range []int{0, 1, 2, 10, 97, 1000, N} {
		want := CountSeq(n)
		for _, parts := range []int{1, 2, 3, 7, 8, 1000, 200_000} {
			if got := CountPar(n, parts); got != want {
				t.Errorf("CountPar(%d, %d) = %d, want %d", n, parts, got, want)
			}
		}
	}
}

// TestChunkCost shows why equal-length chunks are not equal work:
// the chunks with bigger numbers take longer.
func TestChunkCost(t *testing.T) {
	for _, parts := range []int{2, 8} {
		var total time.Duration
		took := make([]time.Duration, parts)
		for i := range parts {
			from, to := N*i/parts, N*(i+1)/parts
			start := time.Now()
			for range 20 {
				sink = countRange(from, to)
			}
			took[i] = time.Since(start) / 20
			total += took[i]
		}
		longest := took[parts-1] // the last chunk is the longest
		t.Logf("parts=%d: chunks %v; total %v, longest %v, best possible speed-up %.2fx",
			parts, took, total, longest, float64(total)/float64(longest))
	}
}

func BenchmarkCountSeq(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sink = CountSeq(N)
	}
}

func BenchmarkCountPar(b *testing.B) {
	for _, parts := range []int{2, 8, 1000, 200_000} {
		b.Run(fmt.Sprintf("parts=%d", parts), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				sink = CountPar(N, parts)
			}
		})
	}
}
