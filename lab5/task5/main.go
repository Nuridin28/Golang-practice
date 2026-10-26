package main

import (
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"slices"
	"time"
)

func rule1() {
	defer fmt.Println("1: deferred first")
	defer fmt.Println("2: deferred second")
	defer fmt.Println("3: deferred third")
	fmt.Println("main body")
}

func rule2() {
	x := 1
	defer fmt.Println("deferred x =", x)
	defer func() { fmt.Println("closure x =", x) }()
	x = 2
	fmt.Println("current x =", x)
}

func double() (result int) {
	defer func() { result *= 2 }()
	return 21
}

func notDouble() int {
	result := 21
	defer func() { result *= 2 }()
	return result
}

func trace(name string) func() {
	fmt.Println("enter", name)
	return func() { fmt.Println("exit", name) }
}

func work() {
	defer trace("work")()
	fmt.Println("working...")
}

type flakyFile struct{}

func (flakyFile) Write(p []byte) (int, error) { return len(p), nil }
func (flakyFile) Close() error                { return errors.New("close grades.csv: no space left on device") }

func save(f io.WriteCloser) (err error) {
	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()
	_, err = io.WriteString(f, "Aruzhan,95\n")
	return err
}

func safeDivide(a, b int) (result int, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		}
	}()
	return a / b, nil
}

func level3() {
	defer fmt.Println("defer in level3")
	panic("something went wrong deep inside")
}

func level2() {
	defer fmt.Println("defer in level2")
	level3()
	fmt.Println("after level3")
}

func level1() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered in level1:", r)
		}
	}()
	level2()
	fmt.Println("after level2")
}

func helper() {
	if r := recover(); r != nil {
		fmt.Println("recovered:", r)
	}
}

func a() {
	defer func() {
		helper()
	}()
	panic("boom A")
}

func b() {
	defer helper()
	panic("boom B")
}

func c() {
	fmt.Println("c: recover outside defer returned", recover())
}

func mustNotPanic(name string, f func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("%s: panic escaped: %v\n", name, r)
		}
	}()
	f()
	fmt.Printf("%s: returned normally\n", name)
}

var openNow, openMax int

type Resource struct{ name string }

func Open(name string) (*Resource, error) {
	if name == "" {
		return nil, errors.New("open: empty name")
	}
	openNow++
	openMax = max(openMax, openNow)
	fmt.Println("open ", name)
	return &Resource{name: name}, nil
}

func (r *Resource) Close() error {
	openNow--
	fmt.Println("close", r.name)
	return nil
}

func processAllBad(names []string) error {
	for _, name := range names {
		r, err := Open(name)
		if err != nil {
			return err
		}
		defer r.Close()
		fmt.Println("  use", r.name)
	}
	return nil
}

func processAllGood(names []string) error {
	for _, name := range names {
		if err := processOne(name); err != nil {
			return err
		}
	}
	return nil
}

func processOne(name string) error {
	r, err := Open(name)
	if err != nil {
		return err
	}
	defer r.Close()
	fmt.Println("  use", r.name)
	return nil
}

func timer(name string) func() {
	start := time.Now()
	return func() { fmt.Printf("%s took %v\n", name, time.Since(start)) }
}

func sortNumbers(n int) {
	defer timer("sort")()
	nums := make([]int, n)
	for i := range nums {
		nums[i] = rand.IntN(1_000_000)
	}
	slices.Sort(nums)
}

func main() {
	fmt.Println("=== Rule 1: LIFO")
	rule1()

	fmt.Println("=== Rule 2: arguments are evaluated at the defer line")
	rule2()

	fmt.Println("=== Rule 3: a deferred closure can change named results")
	fmt.Println(double(), notDouble())

	fmt.Println("=== trace")
	work()

	fmt.Println("=== save: the Close error is not lost")
	fmt.Println("save:", save(flakyFile{}))

	fmt.Println("=== safeDivide")
	fmt.Println(safeDivide(10, 2))
	fmt.Println(safeDivide(1, 0))

	fmt.Println("=== how a panic travels")
	level1()
	fmt.Println("main continues normally")

	fmt.Println("=== which one recovers: a or b?")
	mustNotPanic("a", a)
	mustNotPanic("b", b)
	c()

	names := []string{"a.csv", "b.csv", "c.csv"}
	fmt.Println("=== processAllBad")
	openMax = 0
	if err := processAllBad(names); err != nil {
		fmt.Println(err)
	}
	fmt.Println("max open at once:", openMax)

	fmt.Println("=== processAllGood")
	openMax = 0
	if err := processAllGood(names); err != nil {
		fmt.Println(err)
	}
	fmt.Println("max open at once:", openMax)

	fmt.Println("=== timer")
	sortNumbers(1_000_000)
}
