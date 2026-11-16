// Package task7: how expensive is a channel?
//
//	go test -v ./task7
//	go test -run '^$' -bench . -benchmem -cpu 1,2 ./task7
package task7

const N = 10_000

// passN sends 0..N-1 one by one over a channel with a buffer of `size`
// and sums them on the receiving side.
func passN(size int) int {
	ch := make(chan int, size)
	go func() {
		for i := 0; i < N; i++ {
			ch <- i
		}
		close(ch)
	}()
	sum := 0
	for v := range ch {
		sum += v
	}
	return sum
}

const Batch = 100

// passBatches sends the same numbers as []int slices of Batch values:
// one channel operation per 100 numbers instead of per number.
//
// Every batch is a NEW slice: after `ch <- b` the receiver owns b, so the
// sender must not touch it again (reusing one buffer would be a data race).
func passBatches() int {
	ch := make(chan []int, 1)
	go func() {
		for i := 0; i < N; i += Batch {
			b := make([]int, 0, Batch)
			for j := i; j < i+Batch && j < N; j++ {
				b = append(b, j)
			}
			ch <- b
		}
		close(ch)
	}()
	sum := 0
	for b := range ch {
		for _, v := range b {
			sum += v
		}
	}
	return sum
}

// passDirect: the same sum without any goroutine or channel (the baseline).
func passDirect() int {
	sum := 0
	for i := 0; i < N; i++ {
		sum += i
	}
	return sum
}
