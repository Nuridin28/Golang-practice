// Task 3: Find the bug (a nil error that is not nil).
package main

import (
	"errors"
	"fmt"
)

type ValidationError struct {
	Field string
}

func (e *ValidationError) Error() string { return "invalid field: " + e.Field }

// validateBuggy is the original version. For a valid age it returns an error
// interface holding (*ValidationError, nil), which is not a nil interface.
func validateBuggy(age int) error {
	var err *ValidationError
	if age < 0 {
		err = &ValidationError{Field: "age"}
	}
	return err
}

// validate follows the rules: the result type is error, and "no error" is a
// literal nil.
func validate(age int) error {
	if age < 0 {
		return &ValidationError{Field: "age"}
	}
	return nil
}

func main() {
	fmt.Println("--- buggy")
	err := validateBuggy(20)
	if err != nil {
		// fmt calls err.Error(), which panics on e.Field (nil receiver).
		// fmt recovers from that panic and prints <nil> for a nil pointer.
		fmt.Println("error:", err)
	}
	fmt.Println("err == nil:", err == nil)
	fmt.Printf("%%T: %T\n", err)
	fmt.Printf("%%v: %v\n", err)
	fmt.Println("pointer inside == nil:", err.(*ValidationError) == nil)

	fmt.Println("--- fixed")
	err = validate(20)
	fmt.Println("validate(20) == nil:", err == nil)
	fmt.Printf("%%T: %T\n", err)

	err = validate(-1)
	var ve *ValidationError
	if errors.As(err, &ve) {
		fmt.Println("validate(-1): field =", ve.Field)
	}

	// errors.As also looks inside wrapped errors, a plain assertion does not.
	wrapped := fmt.Errorf("register user: %w", validate(-1))
	_, ok := wrapped.(*ValidationError)
	fmt.Println("wrapped:", wrapped)
	fmt.Println("  assertion ok:", ok)
	if errors.As(wrapped, &ve) {
		fmt.Println("  errors.As field =", ve.Field)
	}
}
