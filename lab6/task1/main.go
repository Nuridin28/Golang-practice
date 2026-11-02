// Task 1: Explore the stack.
//
//	go run ./task1 depth       depth(1_000_000) with pad [64]byte
//	go run ./task1 depth1024   the same with pad [1024]byte
//	go run ./task1 noread      pad is written but never read
//	go run ./task1 forever     endless recursion, with an attempt to recover
//
// Frame sizes: go build -o /dev/null -gcflags=-S ./task1 2>&1 | grep 'TEXT.*depth'
package main

import (
	"fmt"
	"os"
)

func depth(n int) int {
	var pad [64]byte // make every frame a bit bigger
	pad[n%len(pad)] = byte(n)
	if n == 0 {
		return 0
	}
	return depth(n-1) + 1 + int(pad[(n+1)%len(pad)]) // reads a 0, but the compiler can't know that
}

func depth1024(n int) int {
	var pad [1024]byte
	pad[n%len(pad)] = byte(n)
	if n == 0 {
		return 0
	}
	return depth1024(n-1) + 1 + int(pad[(n+1)%len(pad)])
}

// depthNoRead is depth without the line that reads pad.
func depthNoRead(n int) int {
	var pad [64]byte
	pad[n%len(pad)] = byte(n)
	if n == 0 {
		return 0
	}
	return depthNoRead(n-1) + 1
}

func forever(n int) int {
	return forever(n+1) + 1
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("usage: go run ./task1 depth|depth1024|noread|forever")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "depth":
		fmt.Println("depth:", depth(1_000_000))
	case "depth1024":
		fmt.Println("depth1024:", depth1024(1_000_000))
	case "noread":
		fmt.Println("noread:", depthNoRead(1_000_000))
	case "forever":
		defer func() {
			fmt.Println("recovered:", recover()) // never printed: stack overflow is fatal
		}()
		fmt.Println(forever(0))
	}
}
