package main

import "fmt"

func MapOperationsDemo() {
	m := map[string]int{"one": 1, "two": 2}
	fmt.Println("read:", m["one"])

	m["two"] = 20
	fmt.Println("update:", m["two"])

	delete(m, "one")
	fmt.Println("after delete:", m)

	if v, ok := m["one"]; ok {
		fmt.Println("one is present:", v)
	} else {
		fmt.Println("one is missing")
	}
	fmt.Println()
}

func ScoresTask() {
	scores := map[string]int{
		"Alice": 90,
		"Bob":   0,
	}

	if _, ok := scores["Bob"]; ok {
		fmt.Println("Bob is in the map")
	} else {
		fmt.Println("Bob is missing")
	}

	if _, ok := scores["Charlie"]; ok {
		fmt.Println("Charlie is in the map")
	} else {
		fmt.Println("Charlie is missing")
	}
	fmt.Println()
}
