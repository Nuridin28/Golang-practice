package main

import "fmt"

func InventoryTask() {
	inventory := map[string]int{
		"apples":  10,
		"bananas": 5,
	}

	fmt.Println("apples in stock:", inventory["apples"])

	inventory["bananas"] = 12

	inventory["oranges"] = 8

	delete(inventory, "apples")

	fmt.Println("updated inventory:", inventory)
	fmt.Println()
}
