package main

import (
	"errors"
	"fmt"
	"runtime"
)

type Task struct {
	Name string
	Run  func() error
}

func SafeRun(tasks []Task) []error {
	var errs []error
	for _, t := range tasks {
		if err := runTask(t); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func runTask(t Task) (err error) {
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if e, ok := r.(error); ok {
			err = fmt.Errorf("task %s: panic: %w", t.Name, e)
		} else {
			err = fmt.Errorf("task %s: panic: %v", t.Name, r)
		}
	}()
	if err := t.Run(); err != nil {
		return fmt.Errorf("task %s: %w", t.Name, err)
	}
	return nil
}

var ErrNoDiskSpace = errors.New("no disk space left")

func demoTasks() []Task {
	return []Task{
		{"backup", func() error {
			fmt.Println("  backup: done")
			return nil
		}},
		{"index", func() error {
			s := []int{1, 2, 3}
			i := 5
			fmt.Println(s[i])
			return nil
		}},
		{"nil-map", func() error {
			var m map[string]int
			m["x"] = 1
			return nil
		}},
		{"config", func() error {
			panic("config is missing")
		}},
		{"upload", func() error {
			return ErrNoDiskSpace
		}},
		{"panic-nil", func() error {
			panic(nil)
		}},
		{"cleanup", func() error {
			fmt.Println("  cleanup: done")
			return nil
		}},
	}
}

func main() {
	errs := SafeRun(demoTasks())
	fmt.Printf("%d errors:\n", len(errs))
	for _, err := range errs {
		fmt.Println(" ", err)
	}

	joined := errors.Join(errs...)
	fmt.Println("errors.Is(joined, ErrNoDiskSpace):", errors.Is(joined, ErrNoDiskSpace))

	var re runtime.Error
	fmt.Println("errors.As(joined, &runtime.Error):", errors.As(joined, &re), "->", re)

	var pn *runtime.PanicNilError
	fmt.Println("errors.As(joined, &pn) (*runtime.PanicNilError):", errors.As(joined, &pn))

	fmt.Println("no failures:", SafeRun([]Task{{"ok", func() error { return nil }}}) == nil)
}
