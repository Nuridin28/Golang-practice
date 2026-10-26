package main

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strconv"
)

var ErrNotFound = errors.New("student not found")

type GradeError struct {
	Student string
	Grade   int
}

func (e *GradeError) Error() string {
	return fmt.Sprintf("invalid grade %d for %s", e.Grade, e.Student)
}

func findStudent(name string) error {
	return ErrNotFound
}

func loadReport(name string) error {
	if err := findStudent(name); err != nil {
		return fmt.Errorf("load report for %q: %w", name, err)
	}
	return nil
}

func handler() error {
	if err := loadReport("Timur"); err != nil {
		return fmt.Errorf("GET /report: %w", err)
	}
	return nil
}

type QueryError struct {
	Query string
	Err   error
}

func (e *QueryError) Error() string { return "query " + e.Query + ": " + e.Err.Error() }
func (e *QueryError) Unwrap() error { return e.Err }

var ErrTimeout = errors.New("timeout")

func main() {
	fmt.Println("--- standard library errors")
	n, err := strconv.Atoi("42a")
	fmt.Println(n, err)
	_, err = os.Open("missing.txt")
	fmt.Println(err)

	fmt.Println("--- sentinel errors are compared by identity, not by text")
	fmt.Println(errors.New("x") == errors.New("x"))

	fmt.Println("--- wrapping chain")
	err = handler()
	fmt.Println(err)
	fmt.Println(err == ErrNotFound)
	fmt.Println(errors.Is(err, ErrNotFound))
	for e := err; e != nil; e = errors.Unwrap(e) {
		fmt.Printf("  %T: %v\n", e, e)
	}

	fmt.Println("--- errors.As")
	var wrapped error = fmt.Errorf("saving: %w", &GradeError{"Dias", 120})
	var ge *GradeError
	if errors.As(wrapped, &ge) {
		fmt.Println("student:", ge.Student, "grade:", ge.Grade)
	}

	fmt.Printf("--- %%v breaks the chain\n")
	lost := fmt.Errorf("load report: %v", ErrNotFound)
	fmt.Println(lost)
	fmt.Println(errors.Is(lost, ErrNotFound))

	fmt.Println("--- own type with Unwrap")
	err = fmt.Errorf("load users: %w", &QueryError{Query: "SELECT * FROM users", Err: ErrTimeout})
	fmt.Println(err)
	fmt.Println(errors.Is(err, ErrTimeout))

	fmt.Println("--- errors.Join")
	joined := errors.Join(ErrNotFound, &GradeError{"Ali", -5})
	fmt.Println(joined)
	fmt.Println(errors.Is(joined, ErrNotFound))
	fmt.Println(errors.Join() == nil)
	fmt.Println(errors.Join(nil, nil) == nil)
	fmt.Println(errors.Unwrap(joined))

	fmt.Println("--- several %w")
	two := fmt.Errorf("two problems: %w; %w", ErrNotFound, ErrTimeout)
	fmt.Println(two)
	fmt.Println(errors.Is(two, ErrNotFound), errors.Is(two, ErrTimeout))

	fmt.Println("--- standard library errors work the same way")
	_, err = os.Open("missing.txt")
	err = fmt.Errorf("config: %w", err)
	fmt.Println(errors.Is(err, fs.ErrNotExist))
	var pe *fs.PathError
	if errors.As(err, &pe) {
		fmt.Println("op:", pe.Op, "path:", pe.Path)
	}
}
