// Task 4: Tour of Go, Exercise: Errors.
package main

import (
	"fmt"
	"math"
)

type ErrNegativeSqrt float64

// float64(e) is required: fmt.Sprint(e) would see that e is an error, call
// e.Error() again, and recurse until the stack overflows.
func (e ErrNegativeSqrt) Error() string {
	return fmt.Sprint("cannot Sqrt negative number: ", float64(e))
}

// Sqrt uses Newton's method (Exercise: Loops and Functions).
func Sqrt(x float64) (float64, error) {
	if x < 0 {
		return 0, ErrNegativeSqrt(x)
	}
	if x == 0 {
		return 0, nil
	}
	z := 1.0
	for i := 0; i < 100; i++ {
		next := z - (z*z-x)/(2*z)
		if math.Abs(next-z) < 1e-12 {
			break
		}
		z = next
	}
	return z, nil
}

func main() {
	fmt.Println(Sqrt(2))
	fmt.Println(Sqrt(-2))
	fmt.Println(Sqrt(0))
	fmt.Println(Sqrt(144))

	if _, err := Sqrt(-9); err != nil {
		fmt.Printf("%T: %v\n", err, err)
	}
}
