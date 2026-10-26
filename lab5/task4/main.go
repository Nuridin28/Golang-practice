package main

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrBadFormat = errors.New(`bad format, want "name, grade"`)

type Record struct {
	Name  string
	Grade int
}

func ParseAll(input string) ([]Record, error) {
	var records []Record
	var errs []error
	for i, line := range strings.Split(input, "\n") {
		n := i + 1
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		name, gradeText, ok := strings.Cut(line, ",")
		name = strings.TrimSpace(name)
		if !ok || name == "" {
			errs = append(errs, fmt.Errorf("line %d: %q: %w", n, line, ErrBadFormat))
			continue
		}
		grade, err := strconv.Atoi(strings.TrimSpace(gradeText))
		if err != nil {
			errs = append(errs, fmt.Errorf("line %d: %w", n, err))
			continue
		}
		records = append(records, Record{Name: name, Grade: grade})
	}
	return records, errors.Join(errs...)
}

func main() {
	input := `Aruzhan, 95
Dias, 8O
Timur
Bota, 99999999999999999999
Ali, 70`

	records, err := ParseAll(input)
	fmt.Println("records:", records)
	fmt.Println("error:")
	fmt.Println(err)

	fmt.Println("--- errors.Is")
	fmt.Println("ErrBadFormat:     ", errors.Is(err, ErrBadFormat))
	fmt.Println("strconv.ErrSyntax:", errors.Is(err, strconv.ErrSyntax))
	fmt.Println("strconv.ErrRange: ", errors.Is(err, strconv.ErrRange))

	fmt.Println("--- errors.As")
	var numErr *strconv.NumError
	if errors.As(err, &numErr) {
		fmt.Printf("first *strconv.NumError: Func=%s Num=%q Err=%v\n", numErr.Func, numErr.Num, numErr.Err)
	}

	fmt.Println("--- every *strconv.NumError")
	if joined, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range joined.Unwrap() {
			if errors.As(e, &numErr) {
				fmt.Printf("Num=%q Err=%v\n", numErr.Num, numErr.Err)
			}
		}
	}

	fmt.Println("--- all lines are correct")
	records, err = ParseAll("Aruzhan, 95\nAli, 70\n")
	fmt.Println("records:", records)
	fmt.Println("err:", err, "| err == nil:", err == nil)
}
