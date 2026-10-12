// Task 3: Compile or not?
// Lines that do not compile are commented out. Uncomment one and run
// `go build .` to see the error.
package main

import "fmt"

type Counter struct{ n int }

func (c Counter) Get() int { return c.n }
func (c *Counter) Inc()    { c.n++ }

func NewCounter() Counter { return Counter{} }

func main() {
	var c Counter
	c.Inc()         // 1 ✅ c is a variable (addressable): Go rewrites it as (&c).Inc()
	(&c).Inc()      // 2 ✅ explicit pointer
	Counter{}.Get() // 3 ✅ Get has a value receiver, so no address is needed
	// Counter{}.Inc()    // 4 ❌ cannot call pointer method Inc on Counter: a struct literal is not addressable
	// NewCounter().Inc() // 5 ❌ cannot call pointer method Inc on Counter: a return value is not addressable

	m := map[string]Counter{"a": {}}
	m["a"].Get() // 6 ✅ value receiver: reading a copy of the map element is fine
	// m["a"].Inc() // 7 ❌ cannot call pointer method Inc on Counter: map elements are not addressable
	// (the map may move its elements when it grows, so &m["a"] is forbidden)

	s := []Counter{{}}
	s[0].Inc() // 8 ✅ slice elements are addressable: (&s[0]).Inc()

	pm := map[string]*Counter{"a": {}}
	pm["a"].Inc() // 9 ✅ the element is already a *Counter, no & needed

	fmt.Println(c.n, s[0].n, pm["a"].n) // 2 1 1
}
