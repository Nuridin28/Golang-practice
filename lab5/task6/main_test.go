package main

import (
	"errors"
	"runtime"
	"testing"
)

func TestSafeRunMessages(t *testing.T) {
	want := []string{
		"task index: panic: runtime error: index out of range [5] with length 3",
		"task nil-map: panic: assignment to entry in nil map",
		"task config: panic: config is missing",
		"task upload: no disk space left",
		"task panic-nil: panic: runtime error: panic called with nil argument",
	}
	errs := SafeRun(demoTasks())
	if len(errs) != len(want) {
		t.Fatalf("got %d errors, want %d: %v", len(errs), len(want), errs)
	}
	for i, err := range errs {
		if err.Error() != want[i] {
			t.Errorf("errs[%d] = %q, want %q", i, err, want[i])
		}
	}
}

func TestSafeRunRunsEveryTask(t *testing.T) {
	ran := 0
	tasks := []Task{
		{"first", func() error { ran++; panic("boom") }},
		{"second", func() error { ran++; return nil }},
		{"third", func() error { ran++; panic("boom") }},
		{"fourth", func() error { ran++; return nil }},
	}
	errs := SafeRun(tasks)
	if ran != len(tasks) {
		t.Errorf("ran %d tasks, want %d", ran, len(tasks))
	}
	if len(errs) != 2 {
		t.Errorf("got %d errors, want 2: %v", len(errs), errs)
	}
}

func TestSentinelInsideJoin(t *testing.T) {
	joined := errors.Join(SafeRun(demoTasks())...)
	if !errors.Is(joined, ErrNoDiskSpace) {
		t.Error("errors.Is(joined, ErrNoDiskSpace) = false, want true")
	}
	var re runtime.Error
	if !errors.As(joined, &re) {
		t.Error("errors.As(joined, &runtime.Error) = false, want true")
	}
}

func TestSafeRunNoErrors(t *testing.T) {
	errs := SafeRun([]Task{{"ok", func() error { return nil }}})
	if errs != nil {
		t.Errorf("SafeRun = %v, want nil", errs)
	}
	if err := errors.Join(errs...); err != nil {
		t.Errorf("errors.Join = %v, want nil", err)
	}
}

func TestSafeRunNilFunc(t *testing.T) {
	errs := SafeRun([]Task{{Name: "empty"}})
	if len(errs) != 1 {
		t.Fatalf("got %d errors, want 1", len(errs))
	}
	var re runtime.Error
	if !errors.As(errs[0], &re) {
		t.Errorf("calling a nil Run: %v, want a runtime.Error", errs[0])
	}
}
