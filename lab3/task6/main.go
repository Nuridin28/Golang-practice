// Task 6: Shapes.
package main

import (
	"fmt"
	"math"
)

type Shape interface {
	Area() float64
	Perimeter() float64
}

type Rectangle struct{ Width, Height float64 }

func (r Rectangle) Area() float64      { return r.Width * r.Height }
func (r Rectangle) Perimeter() float64 { return 2 * (r.Width + r.Height) }

func (r *Rectangle) Scale(k float64) {
	r.Width *= k
	r.Height *= k
}

type Circle struct{ Radius float64 }

func (c *Circle) Area() float64      { return math.Pi * c.Radius * c.Radius }
func (c *Circle) Perimeter() float64 { return 2 * math.Pi * c.Radius }

// Compile-time checks.
var (
	_ Shape = Rectangle{}
	_ Shape = (*Circle)(nil)
)

func main() {
	r := Rectangle{Width: 3, Height: 4}
	r.Scale(2)

	shapes := []Shape{
		r,
		&Circle{Radius: 1},
		// Circle{Radius: 1},
		// compile error: Circle does not implement Shape (method Area has pointer receiver)
		// Area and Perimeter belong only to *Circle, and a value stored in an
		// interface is not addressable, so Go can't take its address.
	}

	for _, s := range shapes {
		fmt.Printf("%-18T area=%7.2f perimeter=%6.2f\n", s, s.Area(), s.Perimeter())
	}
}
