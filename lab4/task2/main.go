// Task 2: Describe anything + participation task (no type switch).
package main

import (
	"errors"
	"fmt"
)

type Celsius float64

func (c Celsius) String() string { return fmt.Sprintf("%.1f°C", float64(c)) }

func describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case int, int64:
		return fmt.Sprintf("integer %v", x) // x is any here: two types in the case
	case string:
		return fmt.Sprintf("string of %d bytes", len(x)) // x is a string here
	case bool:
		return fmt.Sprintf("bool %t", x)
	case []int:
		return fmt.Sprintf("slice with %d ints", len(x)) // x is a []int here
	case map[string]int:
		return fmt.Sprintf("map with %d keys", len(x))
	case error:
		return "error: " + x.Error()
	case fmt.Stringer:
		return "Stringer: " + x.String()
	default:
		return fmt.Sprintf("unknown type %T", x)
	}
}

// describeAssert is the participation task: the same logic with only
// "comma ok" assertions. Order of the checks matters exactly like the order
// of the cases above.
func describeAssert(v any) string {
	if v == nil {
		return "nil"
	}
	if x, ok := v.(int); ok {
		return fmt.Sprintf("integer %v", x)
	}
	if x, ok := v.(int64); ok {
		return fmt.Sprintf("integer %v", x)
	}
	if x, ok := v.(string); ok {
		return fmt.Sprintf("string of %d bytes", len(x))
	}
	if x, ok := v.(bool); ok {
		return fmt.Sprintf("bool %t", x)
	}
	if x, ok := v.([]int); ok {
		return fmt.Sprintf("slice with %d ints", len(x))
	}
	if x, ok := v.(map[string]int); ok {
		return fmt.Sprintf("map with %d keys", len(x))
	}
	if x, ok := v.(error); ok {
		return "error: " + x.Error()
	}
	if x, ok := v.(fmt.Stringer); ok {
		return "Stringer: " + x.String()
	}
	return fmt.Sprintf("unknown type %T", v)
}

func main() {
	tests := []struct {
		v        any
		expected string
	}{
		{nil, "nil"},
		{7, "integer 7"},
		{int64(7), "integer 7"},
		{"привет", "string of 12 bytes"},
		{[]int{1, 2, 3}, "slice with 3 ints"},
		{errors.New("disk full"), "error: disk full"},
		{Celsius(21.5), "Stringer: 21.5°C"},
		{3.14, "unknown type float64"},
		{struct{}{}, "unknown type struct {}"},
		{true, "bool true"},
		{map[string]int{"go": 1, "rust": 2}, "map with 2 keys"},
	}

	for _, tt := range tests {
		got := describe(tt.v)
		mark := "ok"
		if got != tt.expected {
			mark = "MISMATCH, expected " + tt.expected
		}
		if describeAssert(tt.v) != got {
			mark += " (describeAssert differs!)"
		}
		fmt.Printf("%-24s %s\n", got, mark)
	}
}
