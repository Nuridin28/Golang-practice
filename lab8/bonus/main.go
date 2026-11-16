// Bonus: select on a closed and on a nil channel.
//
//	go run ./bonus            closed + default (1000 times), nil + default (1000 times)
//	go run ./bonus nodefault  nil channel without default: fatal error: deadlock
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "nodefault" {
		var ch chan int
		select {
		case v, ok := <-ch: // never ready, and there's nothing else: select {} in disguise
			fmt.Println("nil:", v, ok)
		}
		return
	}

	count := map[string]int{}
	for i := 0; i < 1000; i++ {
		ch := make(chan int)
		close(ch)
		select {
		case v, ok := <-ch: // a closed channel is ALWAYS ready: 0, false
			count[fmt.Sprint("closed: ", v, " ", ok)]++
		default:
			count["default"]++
		}
	}
	fmt.Println("closed channel + default:", count)

	count = map[string]int{}
	for i := 0; i < 1000; i++ {
		var ch chan int // ch = nil
		select {
		case v, ok := <-ch: // a nil channel is NEVER ready
			count[fmt.Sprint("nil: ", v, " ", ok)]++
		default:
			count["default"]++
		}
	}
	fmt.Println("nil channel + default:   ", count)
}
