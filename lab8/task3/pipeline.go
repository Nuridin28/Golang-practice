package main

// Every stage owns its output channel: it creates it, it is the only writer,
// and it closes it (defer close) when its input is exhausted. Callers only get
// a <-chan int, so they cannot send or close by mistake.
//
// A stage knows nothing about its neighbours: it reads `in` until it is closed
// and closes `out`. That's why adding filterEven (3.3) changed zero lines in
// the other stages — only the line in main that wires the stages together.

// generator: stage 1.
func generator(nums ...int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for _, n := range nums {
			out <- n
		}
	}()
	return out
}

// square: stage 2.
func square(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			out <- n * n
		}
	}()
	return out
}

// filterEven: the stage added in 3.3. Same shape as square.
func filterEven(in <-chan int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for n := range in {
			if n%2 == 0 {
				out <- n
			}
		}
	}()
	return out
}

// sum: the last stage (the sink). It doesn't produce a channel, so it has
// nothing to close; it returns when `in` is closed.
func sum(in <-chan int) int {
	s := 0
	for n := range in {
		s += n
	}
	return s
}
