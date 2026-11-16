package task7

import (
	"fmt"
	"testing"
)

const want = N * (N - 1) / 2

var sizes = []int{0, 1, 100, 10_000}

func TestPassN(t *testing.T) {
	for _, size := range []int{0, 1, 2, 7, 100, N - 1, N, 2 * N} {
		if got := passN(size); got != want {
			t.Errorf("passN(%d) = %d, want %d", size, got, want)
		}
	}
}

func TestPassBatches(t *testing.T) {
	if got := passBatches(); got != want {
		t.Errorf("passBatches() = %d, want %d", got, want)
	}
	if got := passDirect(); got != want {
		t.Errorf("passDirect() = %d, want %d", got, want)
	}
}

var sink int

func BenchmarkPassN(b *testing.B) {
	for _, size := range sizes {
		b.Run(fmt.Sprintf("buf=%d", size), func(b *testing.B) {
			for b.Loop() {
				sink = passN(size)
			}
		})
	}
}

func BenchmarkPassBatches(b *testing.B) {
	for b.Loop() {
		sink = passBatches()
	}
}

func BenchmarkPassDirect(b *testing.B) {
	for b.Loop() {
		sink = passDirect()
	}
}
