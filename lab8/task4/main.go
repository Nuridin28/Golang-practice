// Task 4: Exercise "Equivalent Binary Trees" from the Tour of Go.
//
//	go run ./task4               Walk prints a tree, Same on equal and different trees,
//	                             then runtime.NumGoroutine() after 1000 calls of each Same version
//
// Tests: go test -race -v ./task4
package main

import (
	"fmt"
	"runtime"
	"time"

	"lab8/task4/tree"
)

// Walk sends all values of t to ch in sorted order, then closes ch.
//
// Where does close go? Not inside the recursion: the first finished subtree
// would close the channel and the next send would panic. The recursion lives
// in walk, and the non-recursive wrapper closes once, after the whole tree.
func Walk(t *tree.Tree, ch chan int) {
	defer close(ch)
	walk(t, ch)
}

func walk(t *tree.Tree, ch chan int) {
	if t == nil {
		return
	}
	walk(t.Left, ch)
	ch <- t.Value
	walk(t.Right, ch)
}

// SameLeaky is the usual Tour solution: it returns at the first difference.
// The two Walk goroutines are then stuck on `ch <-` forever: a leak.
func SameLeaky(t1, t2 *tree.Tree) bool {
	ch1, ch2 := make(chan int), make(chan int)
	go Walk(t1, ch1)
	go Walk(t2, ch2)
	for {
		v1, ok1 := <-ch1
		v2, ok2 := <-ch2
		if ok1 != ok2 || v1 != v2 {
			return false // the walkers may still have values to send: nobody receives them
		}
		if !ok1 {
			return true
		}
	}
}

// WalkDone is Walk with a done channel: every send also waits on <-done,
// so the walker stops as soon as the reader gives up.
func WalkDone(t *tree.Tree, ch chan<- int, done <-chan struct{}) {
	defer close(ch)
	var walk func(t *tree.Tree) bool
	walk = func(t *tree.Tree) bool {
		if t == nil {
			return true
		}
		if !walk(t.Left) {
			return false
		}
		select {
		case ch <- t.Value:
		case <-done:
			return false // stop the whole recursion
		}
		return walk(t.Right)
	}
	walk(t)
}

// Same reports whether t1 and t2 contain the same values. It doesn't assume
// 10 values per tree (it reads until both channels are closed), and it doesn't
// leak: on return, close(done) tells both walkers to stop.
func Same(t1, t2 *tree.Tree) bool {
	done := make(chan struct{})
	defer close(done)
	ch1, ch2 := make(chan int), make(chan int)
	go WalkDone(t1, ch1, done)
	go WalkDone(t2, ch2, done)
	for {
		v1, ok1 := <-ch1
		v2, ok2 := <-ch2
		if ok1 != ok2 || v1 != v2 {
			return false
		}
		if !ok1 {
			return true
		}
	}
}

func main() {
	ch := make(chan int)
	t := tree.New(1)
	go Walk(t, ch)
	fmt.Println("tree:", t)
	fmt.Print("Walk: ")
	for v := range ch {
		fmt.Print(v, " ")
	}
	fmt.Println()

	fmt.Println("Same(New(1), New(1)):", Same(tree.New(1), tree.New(1)))
	fmt.Println("Same(New(1), New(2)):", Same(tree.New(1), tree.New(2)))

	for _, v := range []struct {
		name string
		same func(a, b *tree.Tree) bool
	}{{"SameLeaky", SameLeaky}, {"Same", Same}} {
		before := runtime.NumGoroutine()
		for i := 0; i < 1000; i++ {
			v.same(tree.New(1), tree.New(1)) // equal: both walkers finish
		}
		time.Sleep(50 * time.Millisecond)
		equal := runtime.NumGoroutine() - before
		for i := 0; i < 1000; i++ {
			v.same(tree.New(1), tree.New(2)) // different: return at the first value
		}
		time.Sleep(50 * time.Millisecond)
		fmt.Printf("%-9s 1000 equal pairs: +%d goroutines, then 1000 different pairs: +%d goroutines\n",
			v.name, equal, runtime.NumGoroutine()-before-equal)
	}
}
