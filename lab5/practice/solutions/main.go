package main

import (
	"errors"
	"fmt"
	"io"
	"iter"
	"strconv"
	"strings"
	"sync"
)

func once(f func()) func() {
	done := false
	return func() {
		if done {
			return
		}
		done = true
		f()
	}
}

func newCounter() (inc func() int, reset func()) {
	n := 0
	inc = func() int {
		n++
		return n
	}
	reset = func() { n = 0 }
	return inc, reset
}

func idGen(prefix string) func() string {
	n := 0
	return func() string {
		n++
		return prefix + "-" + strconv.Itoa(n)
	}
}

func retry(attempts int, f func() error) error {
	var errs []error
	for i := 1; i <= attempts; i++ {
		err := f()
		if err == nil {
			return nil
		}
		errs = append(errs, fmt.Errorf("attempt %d: %w", i, err))
	}
	return errors.Join(errs...)
}

func withRecover(f func()) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if e, ok := r.(error); ok {
			err = fmt.Errorf("panic: %w", e)
		} else {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	f()
	return nil
}

func Must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func closeAll(cs ...io.Closer) error {
	var errs []error
	for _, c := range cs {
		if err := c.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

type fakeCloser struct {
	name string
	err  error
}

func (c fakeCloser) Close() error {
	fmt.Println("close", c.name)
	return c.err
}

type LineError struct {
	Line int
	Err  error
}

func (e *LineError) Error() string { return "line " + strconv.Itoa(e.Line) + ": " + e.Err.Error() }
func (e *LineError) Unwrap() error { return e.Err }

var ErrBadFormat = errors.New("bad format")

func parseGrades(input string) ([]int, error) {
	var grades []int
	var errs []error
	for i, line := range strings.Split(input, "\n") {
		_, text, ok := strings.Cut(line, ",")
		if !ok {
			errs = append(errs, &LineError{Line: i + 1, Err: ErrBadFormat})
			continue
		}
		g, err := strconv.Atoi(strings.TrimSpace(text))
		if err != nil {
			errs = append(errs, &LineError{Line: i + 1, Err: err})
			continue
		}
		grades = append(grades, g)
	}
	return grades, errors.Join(errs...)
}

type FieldError struct {
	Field string
	Msg   string
}

func (e *FieldError) Error() string { return e.Field + ": " + e.Msg }

type Student struct {
	Name  string
	Grade int
}

func validate(s Student) error {
	var errs []error
	if s.Name == "" {
		errs = append(errs, &FieldError{"name", "is empty"})
	}
	if s.Grade < 0 || s.Grade > 100 {
		errs = append(errs, &FieldError{"grade", "must be 0..100, got " + strconv.Itoa(s.Grade)})
	}
	return errors.Join(errs...)
}

type Handler func(string) string

func logging(next Handler) Handler {
	return func(s string) string {
		fmt.Printf("  -> %q\n", s)
		out := next(s)
		fmt.Printf("  <- %q\n", out)
		return out
	}
}

func exclaim(next Handler) Handler {
	return func(s string) string { return next(s) + "!" }
}

func chain(h Handler, mws ...func(Handler) Handler) Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}

func fibSeq(n int) iter.Seq[int] {
	return func(yield func(int) bool) {
		a, b := 0, 1
		for i := 0; i < n; i++ {
			if !yield(a) {
				return
			}
			a, b = b, a+b
		}
	}
}

type errWriter struct {
	w   io.Writer
	err error
}

func (ew *errWriter) write(s string) {
	if ew.err != nil {
		return
	}
	_, ew.err = io.WriteString(ew.w, s)
}

type limitedWriter struct{ left int }

func (l *limitedWriter) Write(p []byte) (int, error) {
	if len(p) > l.left {
		return 0, errors.New("disk full")
	}
	l.left -= len(p)
	return len(p), nil
}

func main() {
	fmt.Println("--- W1 once")
	hello := once(func() { fmt.Println("hello (only once)") })
	hello()
	hello()
	std := sync.OnceFunc(func() { fmt.Println("sync.OnceFunc (only once)") })
	std()
	std()

	fmt.Println("--- W2 newCounter")
	inc, reset := newCounter()
	fmt.Println(inc(), inc(), inc())
	reset()
	fmt.Println(inc())

	fmt.Println("--- W3 idGen")
	users, orders := idGen("user"), idGen("order")
	fmt.Println(users(), users(), orders(), users())

	fmt.Println("--- W4 retry")
	calls := 0
	err := retry(3, func() error {
		calls++
		if calls < 2 {
			return errors.New("timeout")
		}
		return nil
	})
	fmt.Println("succeeded on call", calls, "err:", err)
	err = retry(3, func() error { return errors.New("connection refused") })
	fmt.Println(err)

	fmt.Println("--- W5 withRecover")
	fmt.Println(withRecover(func() { panic("boom") }))
	err = withRecover(func() {
		var m map[string]int
		m["x"] = 1
	})
	fmt.Println(err)
	fmt.Println(withRecover(func() {}))

	fmt.Println("--- W6 Must")
	n := Must(strconv.Atoi("42"))
	fmt.Println(n)
	fmt.Println(withRecover(func() { Must(strconv.Atoi("4x2")) }))

	fmt.Println("--- W7 closeAll")
	err = closeAll(
		fakeCloser{"db", nil},
		fakeCloser{"file", errors.New("close file: disk full")},
		fakeCloser{"conn", errors.New("close conn: broken pipe")},
	)
	fmt.Println(err)

	fmt.Println("--- W8 LineError")
	grades, err := parseGrades("Aruzhan, 95\nTimur\nDias, 8O")
	fmt.Println(grades)
	fmt.Println(err)
	var le *LineError
	if errors.As(err, &le) {
		fmt.Println("first bad line:", le.Line, "| bad format:", errors.Is(err, ErrBadFormat))
	}

	fmt.Println("--- W9 validate")
	err = validate(Student{Name: "", Grade: 120})
	fmt.Println(err)
	var fe *FieldError
	if errors.As(err, &fe) {
		fmt.Println("first field:", fe.Field)
	}
	fmt.Println("valid:", validate(Student{"Ali", 70}) == nil)

	fmt.Println("--- W10 middleware")
	h := chain(strings.ToUpper, logging, exclaim)
	fmt.Println(h("go"))

	fmt.Println("--- W11 iterator")
	for v := range fibSeq(10) {
		if v > 20 {
			break
		}
		fmt.Print(v, " ")
	}
	fmt.Println()

	fmt.Println("--- W12 errWriter")
	ew := &errWriter{w: &limitedWriter{left: 10}}
	ew.write("12345")
	ew.write("67890")
	ew.write("overflow")
	ew.write("never written")
	fmt.Println("err:", ew.err)
}
