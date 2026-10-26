package main

import (
	"errors"
	"fmt"
)

var ErrNotFound = errors.New("student not found")

type GradeError struct {
	Student string
	Grade   int
}

func (e *GradeError) Error() string {
	return fmt.Sprintf("invalid grade %d for %s", e.Grade, e.Student)
}

type GradeBook struct {
	grades map[string]int
}

func NewGradeBook() *GradeBook {
	return &GradeBook{grades: make(map[string]int)}
}

func (gb *GradeBook) Get(name string) (int, error) {
	g, ok := gb.grades[name]
	if !ok {
		return 0, ErrNotFound
	}
	return g, nil
}

func (gb *GradeBook) Set(name string, grade int) error {
	if grade < 0 || grade > 100 {
		return &GradeError{Student: name, Grade: grade}
	}
	gb.grades[name] = grade
	return nil
}

func AddBonus(gb *GradeBook, name string, bonus int) error {
	g, err := gb.Get(name)
	if err != nil {
		return fmt.Errorf("add bonus for %q: %w", name, err)
	}
	if err := gb.Set(name, g+bonus); err != nil {
		return fmt.Errorf("add bonus %+d for %q: %w", bonus, name, err)
	}
	return nil
}

func addBonusLossy(gb *GradeBook, name string, bonus int) error {
	g, err := gb.Get(name)
	if err != nil {
		return fmt.Errorf("add bonus for %q: %v", name, err)
	}
	if err := gb.Set(name, g+bonus); err != nil {
		return fmt.Errorf("add bonus %+d for %q: %w", bonus, name, err)
	}
	return nil
}

func report(err error) {
	var ge *GradeError
	switch {
	case err == nil:
		fmt.Println("OK")
	case errors.Is(err, ErrNotFound):
		fmt.Println("404:", err)
	case errors.As(err, &ge):
		fmt.Printf("400: %v (student: %s, grade: %d)\n", err, ge.Student, ge.Grade)
	default:
		fmt.Println("500:", err)
	}
}

func main() {
	gb := NewGradeBook()
	for name, g := range map[string]int{"Aruzhan": 90, "Dias": 95, "Bota": 40} {
		if err := gb.Set(name, g); err != nil {
			panic(err)
		}
	}

	type call struct {
		name  string
		bonus int
	}
	calls := []call{{"Aruzhan", 5}, {"Timur", 5}, {"Dias", 10}, {"Bota", -50}}

	fmt.Printf("--- AddBonus (%%w)\n")
	for _, c := range calls {
		fmt.Printf("%-22s ", fmt.Sprintf("AddBonus(%s, %+d)", c.name, c.bonus))
		report(AddBonus(gb, c.name, c.bonus))
	}
	fmt.Printf("%-22s ", "unexpected error")
	report(errors.New("grade book storage is unavailable"))

	g, _ := gb.Get("Dias")
	fmt.Println("Dias after the failed Set:", g)

	fmt.Printf("--- addBonusLossy (%%v for the Get error)\n")
	for _, c := range calls[1:3] {
		fmt.Printf("%-22s ", fmt.Sprintf("AddBonus(%s, %+d)", c.name, c.bonus))
		report(addBonusLossy(gb, c.name, c.bonus))
	}
}
