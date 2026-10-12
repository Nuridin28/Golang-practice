// Task 1: Student records.
package main

import "fmt"

type Student struct {
	Name    string
	ID      int
	GPA     float64
	Courses []string
}

// addCourse gets a copy of the student, so append only changes the copy's
// slice header and the caller's Courses stays the same.
func addCourse(s Student, course string) {
	s.Courses = append(s.Courses, course)
}

func addCoursePtr(s *Student, course string) {
	s.Courses = append(s.Courses, course)
}

func main() {
	var s1 Student
	s2 := Student{Name: "Aruzhan", ID: 1, GPA: 3.8}
	s3 := Student{"Dias", 2, 3.2, []string{"Go", "Databases"}}
	s4 := &Student{Name: "Timur"}
	s4.GPA = 3.5

	fmt.Printf("s1: %+v\n", s1)
	fmt.Printf("s2: %+v\n", s2)
	fmt.Printf("s3: %+v\n", s3)
	fmt.Printf("s4: %+v\n", s4)

	fmt.Printf("%%v : %v\n", s3)
	fmt.Printf("%%+v: %+v\n", s3)
	fmt.Printf("%%#v: %#v\n", s3)
	fmt.Println("s1.Courses == nil:", s1.Courses == nil)

	addCourse(s2, "Algorithms")
	fmt.Printf("%+v\n", s2)

	addCoursePtr(&s2, "Algorithms")
	fmt.Printf("%+v\n", s2)
}
