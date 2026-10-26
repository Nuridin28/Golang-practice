package task7

import (
	"errors"
	"fmt"
	"sync"
)

var ErrNegative = errors.New("negative value")

//go:noinline
func checkErr(n int) (int, error) {
	if n < 0 {
		return 0, ErrNegative
	}
	return n, nil
}

//go:noinline
func checkPanic(n int) int {
	if n < 0 {
		panic(ErrNegative)
	}
	return n
}

func safeCheck(n int) (v int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	return checkPanic(n), nil
}

//go:noinline
func checkWrapped(n int) (int, error) {
	v, err := checkErr(n)
	if err != nil {
		return 0, fmt.Errorf("check %d: %w", n, err)
	}
	return v, nil
}

type Counter struct {
	mu sync.Mutex
	n  int
}

func (c *Counter) IncNoDefer() { c.mu.Lock(); c.n++; c.mu.Unlock() }
func (c *Counter) IncDefer()   { c.mu.Lock(); defer c.mu.Unlock(); c.n++ }
