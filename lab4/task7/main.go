// Task 7: Tour of Go, Generic types.
package main

import (
	"fmt"
	"strings"
)

// List represents a singly-linked list that holds values of any type.
type List[T any] struct {
	next *List[T]
	val  T
}

// Push adds v to the end and returns the head, like append:
// always write l = l.Push(v). A method with a pointer receiver can be called
// on a nil *List, but it can't change the caller's variable, so for an empty
// list the new node has to be returned.
func (l *List[T]) Push(v T) *List[T] {
	node := &List[T]{val: v}
	if l == nil {
		return node
	}
	last := l
	for last.next != nil {
		last = last.next
	}
	last.next = node
	return l
}

func (l *List[T]) Len() int {
	n := 0
	for ; l != nil; l = l.next {
		n++
	}
	return n
}

func (l *List[T]) String() string {
	var sb strings.Builder
	sb.WriteByte('[')
	for n := l; n != nil; n = n.next {
		if n != l {
			sb.WriteByte(' ')
		}
		fmt.Fprint(&sb, n.val)
	}
	sb.WriteByte(']')
	return sb.String()
}

func main() {
	var nums *List[int]
	fmt.Println(nums, nums.Len())

	for i := 1; i <= 3; i++ {
		nums = nums.Push(i * 10)
	}
	fmt.Println(nums, nums.Len())

	var words *List[string]
	words = words.Push("go").Push("is").Push("fun")
	fmt.Println(words, words.Len())
}
