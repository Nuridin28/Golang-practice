// Package task7 counts primes with and without goroutines.
//
//	go test -v ./task7
//	go test -run '^$' -bench . -benchmem -cpu 1,2,4,10 ./task7
package task7

import "sync"

// isPrime: trial division by every d with d*d <= n.
// A prime p costs about sqrt(p) steps, so big numbers are harder to check.
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for d := 2; d*d <= n; d++ {
		if n%d == 0 {
			return false
		}
	}
	return true
}

// countRange counts the primes in [from, to).
func countRange(from, to int) int {
	count := 0
	for n := from; n < to; n++ {
		if isPrime(n) {
			count++
		}
	}
	return count
}

// CountSeq counts the primes below n without goroutines.
func CountSeq(n int) int {
	return countRange(0, n)
}

// CountPar splits [0, n) into `parts` chunks of (almost) equal length
// and counts every chunk in its own goroutine.
func CountPar(n, parts int) int {
	results := make([]int, parts)
	var wg sync.WaitGroup
	for i := 0; i < parts; i++ {
		from, to := n*i/parts, n*(i+1)/parts // no gaps, the last chunk ends at n
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i] = countRange(from, to) // every goroutine writes only its own element
		}()
	}
	wg.Wait()
	total := 0
	for _, r := range results {
		total += r
	}
	return total
}
