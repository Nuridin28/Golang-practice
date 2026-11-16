// Package tree is a local copy of the idea behind golang.org/x/tour/tree,
// so the lab builds without downloading anything.
//
// New(k) builds a random binary search tree holding k, 2k, ..., 10k. Two calls
// with the same k give different shapes but the same sorted values.
package tree

import (
	"fmt"
	"math/rand/v2"
)

type Tree struct {
	Left  *Tree
	Value int
	Right *Tree
}

// New returns a new, random binary tree holding the values k, 2k, ..., 10k.
func New(k int) *Tree {
	var t *Tree
	for _, v := range rand.Perm(10) {
		t = insert(t, (1+v)*k)
	}
	return t
}

// FromValues builds a search tree from the values in the given order
// (handy in tests for trees of different sizes).
func FromValues(vals ...int) *Tree {
	var t *Tree
	for _, v := range vals {
		t = insert(t, v)
	}
	return t
}

func insert(t *Tree, v int) *Tree {
	if t == nil {
		return &Tree{nil, v, nil}
	}
	if v < t.Value {
		t.Left = insert(t.Left, v)
	} else {
		t.Right = insert(t.Right, v)
	}
	return t
}

func (t *Tree) String() string {
	if t == nil {
		return "()"
	}
	s := ""
	if t.Left != nil {
		s += t.Left.String() + " "
	}
	s += fmt.Sprint(t.Value)
	if t.Right != nil {
		s += " " + t.Right.String()
	}
	return "(" + s + ")"
}
