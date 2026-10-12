// Task 2: University (composition and embedding) + participation task.
package main

import "fmt"

type Person struct {
	Name  string
	Email string
}

func (p Person) Contact() string {
	return p.Name + " <" + p.Email + ">"
}

type Student struct {
	Person
	GPA float64
}

type Teacher struct {
	Person
	Department string
}

// Contact shadows the promoted Person.Contact.
func (t Teacher) Contact() string {
	return t.Person.Contact() + ", " + t.Department + " department"
}

// Participation task: two embedded types with the same field ID.

type User struct{ ID int }
type Order struct{ ID int }

type UserOrder struct {
	User
	Order
}

func main() {
	s := Student{Person: Person{Name: "Aruzhan", Email: "aruzhan@uni.kz"}, GPA: 3.8}
	t := Teacher{Person: Person{Name: "Bakytzhan", Email: "bakytzhan@uni.kz"}, Department: "Computer Science"}

	fmt.Println(s.Contact())
	fmt.Println(t.Contact())
	fmt.Println(t.Person.Contact())

	uo := UserOrder{User: User{ID: 7}, Order: Order{ID: 42}}

	// fmt.Println(uo.ID)
	// compile error: ambiguous selector uo.ID
	// User.ID and Order.ID are at the same depth and UserOrder has no ID of
	// its own, so the compiler can't pick one.

	fmt.Println("User ID:", uo.User.ID, "Order ID:", uo.Order.ID)
}
