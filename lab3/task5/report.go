// Package task5 compares value and pointer receivers (Task 5).
package task5

// Report is deliberately large: 1024 * 8 bytes = 8 KB.
type Report struct {
	Title  string
	Values [1024]float64
}

//go:noinline
func (r Report) FirstValue() float64 { return r.Values[0] }

//go:noinline
func (r *Report) FirstPointer() float64 { return r.Values[0] }

type Point struct{ X, Y float64 }

//go:noinline
func (p Point) XValue() float64 { return p.X }

//go:noinline
func (p *Point) XPointer() float64 { return p.X }
