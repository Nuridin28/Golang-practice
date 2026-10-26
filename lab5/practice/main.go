package main

import (
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"runtime"
	"slices"
	"strconv"
	"sync"
	"time"
)

type Greeter struct{ name string }

func (g Greeter) Hello() { fmt.Println("hello,", g.name) }

func ex01() {
	g := Greeter{"Aruzhan"}
	f := g.Hello
	g.name = "Dias"
	f()
	g.Hello()
}

func pair() (inc func(), get func() int) {
	n := 0
	inc = func() { n++ }
	get = func() int { return n }
	return
}

func ex02() {
	inc, get := pair()
	inc()
	inc()
	fmt.Println(get())
	inc2, get2 := pair()
	inc2()
	fmt.Println(get(), get2())
}

func adder(base int) func(int) int {
	return func(n int) int {
		base += n
		return base
	}
}

func ex03() {
	a := adder(10)
	fmt.Println(a(1), a(1), adder(10)(1))
}

func ex04() {
	x := 1
	f := func() int { return x * 10 }
	g := func(x int) int { return x * 10 }
	x = 5
	fmt.Println(f(), g(2))
}

func appendOne(nums ...int) []int {
	nums[0] = 100
	return append(nums, 1)
}

func ex05() {
	s := []int{1, 2, 3}
	r := appendOne(s...)
	fmt.Println(s, r)
	t := appendOne(1, 2, 3)
	fmt.Println(t)
}

type Counter struct{ n int }

func ex06() {
	c := Counter{}
	inc := func() { c.n++ }
	inc()
	inc()
	d := c
	inc()
	fmt.Println(c.n, d.n)
}

func ex07() {
	ops := map[string]func(int, int) int{
		"+": func(a, b int) int { return a + b },
	}
	f := ops["-"]
	fmt.Println(f == nil)
	fmt.Println(f(1, 2))
}

func deferArg() {
	for i := 0; i < 3; i++ {
		defer fmt.Print(i, " ")
	}
}

func deferClosure() {
	for i := 0; i < 3; i++ {
		defer func() { fmt.Print(i, " ") }()
	}
}

func ex08() {
	deferArg()
	fmt.Println()
	deferClosure()
	fmt.Println()
}

func ex09() {
	names := []string{"a", "b", "c"}
	var ptrs []*string
	for _, n := range names {
		ptrs = append(ptrs, &n)
	}
	for _, p := range ptrs {
		fmt.Print(*p, " ")
	}
	fmt.Println()
}

func ex10() {
	for i := 0; i < 5; i++ {
		fmt.Print(i, " ")
		i++
	}
	fmt.Println()
}

func ex11() {
	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		got []int
	)
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			got = append(got, i)
			mu.Unlock()
		}()
	}
	wg.Wait()
	slices.Sort(got)
	fmt.Println(got)
}

func sumDeferred() (n int) {
	for i := 1; i <= 3; i++ {
		defer func() { n += i }()
	}
	return 10
}

func ex12() {
	fmt.Println(sumDeferred())
}

func ex13() {
	e1 := errors.New("boom")
	e2 := errors.New("boom")
	fmt.Println(e1 == e2, e1.Error() == e2.Error(), errors.Is(e1, e2))
}

func ex14() {
	base := errors.New("base")
	w1 := fmt.Errorf("l1: %w", base)
	w2 := fmt.Errorf("l2: %w", w1)
	fmt.Println(w2)
	fmt.Println(errors.Is(w2, w1), errors.Is(w1, w2))
	fmt.Println(errors.Unwrap(w2) == w1, errors.Unwrap(errors.Unwrap(w2)) == base, errors.Unwrap(base))
}

func ex15() {
	base := errors.New("base")
	a := fmt.Errorf("a: %v", base)
	b := fmt.Errorf("b: %w", a)
	fmt.Println(b)
	fmt.Println(errors.Is(b, base), errors.Is(b, a))
}

type MyErr struct{}

func (*MyErr) Error() string { return "my error" }

func find(ok bool) error {
	var e *MyErr
	if !ok {
		e = &MyErr{}
	}
	return e
}

func ex16() {
	err := find(true)
	fmt.Println(err == nil, err)
}

type ValErr struct{ Code int }

func (e ValErr) Error() string { return "code " + strconv.Itoa(e.Code) }

func ex17() {
	err := fmt.Errorf("wrap: %w", ValErr{404})
	var v ValErr
	var p *ValErr
	fmt.Println(errors.As(err, &v), v.Code)
	fmt.Println(errors.As(err, &p))
}

func ex18() {
	fmt.Println(errors.Join(nil, nil) == nil)
	e := errors.Join(nil, errors.New("x"), nil)
	fmt.Printf("%q\n", e.Error())
	var errs []error
	fmt.Println(errors.Join(errs...) == nil, len(errs))
}

func ex19() {
	var err error
	fmt.Println(errors.Is(err, nil), errors.Is(errors.New("x"), nil))
	var pe *fs.PathError
	fmt.Println(errors.As(err, &pe))
}

func ex20() {
	var cause error
	err := fmt.Errorf("load: %w", cause)
	fmt.Println(err, err == nil, errors.Unwrap(err) == nil)
}

var ErrClient = errors.New("client error")

type HTTPError struct{ Code int }

func (e *HTTPError) Error() string { return "http " + strconv.Itoa(e.Code) }

func (e *HTTPError) Is(target error) bool {
	return target == ErrClient && e.Code >= 400 && e.Code < 500
}

func ex21() {
	e1 := fmt.Errorf("get: %w", &HTTPError{404})
	e2 := fmt.Errorf("get: %w", &HTTPError{503})
	fmt.Println(errors.Is(e1, ErrClient), errors.Is(e2, ErrClient))
	fmt.Println(errors.Is(e1, &HTTPError{404}))
}

func ex22() {
	err := errors.Join(
		fmt.Errorf("first: %w", &fs.PathError{Op: "open", Path: "a.txt", Err: fs.ErrNotExist}),
		&fs.PathError{Op: "read", Path: "b.txt", Err: fs.ErrPermission},
	)
	var pe *fs.PathError
	errors.As(err, &pe)
	fmt.Println(pe.Path, errors.Is(err, fs.ErrPermission))
}

type BadErr struct{ msg *string }

func (b BadErr) Error() string { return *b.msg }

func ex23() {
	var err error = BadErr{}
	fmt.Println(err)
	fmt.Println("program continues")
}

func load() (err error) {
	defer func() { fmt.Println("deferred sees:", err) }()
	if _, err := strconv.Atoi("x"); err != nil {
		fmt.Println("inner:", err)
	}
	return nil
}

func ex24() {
	fmt.Println("returned:", load())
}

func blogA() {
	i := 0
	defer fmt.Println(i)
	i++
}

func blogB() {
	for i := 0; i < 4; i++ {
		defer fmt.Print(i)
	}
}

func blogC() (i int) {
	defer func() { i++ }()
	return 1
}

func ex25() {
	blogA()
	blogB()
	fmt.Println()
	fmt.Println(blogC())
}

func un(s string) { fmt.Println("leave:", s) }

func tr(s string) string {
	fmt.Println("enter:", s)
	return s
}

func traceA() {
	defer un(tr("a"))
	fmt.Println("in a")
}

func traceB() {
	defer un(tr("b"))
	fmt.Println("in b")
	traceA()
}

func ex26() {
	traceB()
}

type Box struct{ v int }

func (b Box) Show()   { fmt.Println("value:", b.v) }
func (b *Box) ShowP() { fmt.Println("pointer:", b.v) }

func ex27() {
	b := Box{1}
	defer b.Show()
	defer b.ShowP()
	b.v = 2
}

func twoDefers() (x int) {
	defer func() { x++ }()
	defer func() { x *= 10 }()
	x = 1
	return x + 1
}

func ex28() {
	fmt.Println(twoDefers())
}

func nilCleanup() {
	var cleanup func()
	defer fmt.Println("A")
	defer cleanup()
	fmt.Println("B")
}

func ex29() {
	defer func() { fmt.Println("recovered:", recover()) }()
	nilCleanup()
}

func ex30() {
	log.SetFlags(0)
	log.SetOutput(os.Stdout)
	defer fmt.Println("deferred")
	log.Panic("bad config")
}

func blogF() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
	fmt.Println("Calling g.")
	blogG(0)
	fmt.Println("Returned normally from g.")
}

func blogG(i int) {
	if i > 3 {
		fmt.Println("Panicking!")
		panic(fmt.Sprintf("%v", i))
	}
	defer fmt.Println("Defer in g", i)
	fmt.Println("Printing in g", i)
	blogG(i + 1)
}

func ex31() {
	blogF()
	fmt.Println("Returned normally from f.")
}

func ex32() {
	defer func() { fmt.Println("second recover:", recover()) }()
	defer func() { fmt.Println("first recover:", recover()) }()
	panic("once")
}

func inner() {
	defer func() {
		if r := recover(); r != nil {
			panic(fmt.Sprintf("wrapped(%v)", r))
		}
	}()
	panic("original")
}

func ex33() {
	defer func() { fmt.Println("outer got:", recover()) }()
	inner()
}

var ErrX = errors.New("x")

func ex34() {
	defer func() {
		r := recover()
		err, ok := r.(error)
		fmt.Println(ok, errors.Is(err, ErrX))
		fmt.Printf("%T\n", r)
	}()
	panic(fmt.Errorf("ctx: %w", ErrX))
}

func ex35() {
	defer func() {
		r := recover()
		re, ok := r.(runtime.Error)
		fmt.Println(ok, re)
	}()
	var arr []int
	idx := 3
	fmt.Println(arr[idx])
}

func ex36() {
	defer func() { fmt.Println("recovered:", recover()) }()
	defer func() { panic("second") }()
	panic("first")
}

type Guard struct{ name string }

func (g Guard) Catch() {
	if r := recover(); r != nil {
		fmt.Println(g.name, "caught:", r)
	}
}

func guarded() {
	defer Guard{"g1"}.Catch()
	panic("boom")
}

func ex37() {
	guarded()
	fmt.Println("after guarded")
}

func unnamed() int {
	defer func() { recover() }()
	panic("x")
}

func named() (s string) {
	defer func() {
		if r := recover(); r != nil {
			s = fmt.Sprint("recovered: ", r)
		}
	}()
	s = "before"
	panic("oops")
}

func ex38() {
	fmt.Println(unnamed())
	fmt.Println(named())
}

func safe(f func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
		}
	}()
	f()
	return nil
}

func ex39() {
	fmt.Println(safe(func() { panic(errors.New("e")) }))
	fmt.Println(safe(func() { panic("text") }))
}

func ex40() {
	defer func() {
		r := recover()
		fmt.Printf("%T: %v\n", r, r)
	}()
	panic(nil)
}

func ex41() {
	defer fmt.Println("deferred print")
	defer recover()
	panic("is it recovered?")
}

func ex42() {
	defer func() { fmt.Println("main recovered:", recover()) }()
	go func() { panic("in goroutine") }()
	time.Sleep(100 * time.Millisecond)
	fmt.Println("after sleep")
}

func exitInside() {
	defer fmt.Println("exitInside deferred")
	os.Exit(0)
}

func ex43() {
	defer fmt.Println("ex43 deferred")
	fmt.Println("start")
	exitInside()
	fmt.Println("end")
}

var exercises = []func(){
	ex01, ex02, ex03, ex04, ex05, ex06, ex07, ex08, ex09, ex10,
	ex11, ex12, ex13, ex14, ex15, ex16, ex17, ex18, ex19, ex20,
	ex21, ex22, ex23, ex24, ex25, ex26, ex27, ex28, ex29, ex30,
	ex31, ex32, ex33, ex34, ex35, ex36, ex37, ex38, ex39, ex40,
	ex41, ex42, ex43,
}

func main() {
	if len(os.Args) != 2 {
		fmt.Printf("usage: go run ./practice N   (N = 1..%d)\n", len(exercises))
		os.Exit(2)
	}
	n, err := strconv.Atoi(os.Args[1])
	if err != nil || n < 1 || n > len(exercises) {
		fmt.Printf("exercise must be a number from 1 to %d\n", len(exercises))
		os.Exit(2)
	}
	exercises[n-1]()
}
