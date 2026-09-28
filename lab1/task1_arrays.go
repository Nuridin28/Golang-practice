package main

import "fmt"

func Task1_1() {
	prices := [5]float32{10.5, 20.25, 30.75, 40.0, 50.125}
	fmt.Println("prices:", prices, "len:", len(prices))
}

func Task1_2() {
	partial := [5]int{1, 2, 3}
	words := [5]string{"go", "is", "fun"}
	sparse := [5]int{0: 100, 4: 500}

	fmt.Println("partial:", partial, "len:", len(partial))
	fmt.Printf("words: %q\n", words)
	fmt.Println("sparse:", sparse)
	fmt.Println()
}

func Task1_3() {
	var grid [3][4]int
	fmt.Println("2D array:", grid)

	grid[0][0] = 1
	grid[1][2] = 7
	grid[2][3] = 9
	fmt.Println("after assignment:", grid)

	matrix := [2][3]int{
		{1, 2, 3},
		{4, 5, 6},
	}
	for _, row := range matrix {
		fmt.Println(" ", row)
	}
	fmt.Println("rows:", len(matrix), "cols:", len(matrix[0]))
}

func ParticipationTask1() {
	var board [3][3]string
	for i := range board {
		for j := range board[i] {
			board[i][j] = "."
		}
	}
	board[0][0] = "X"
	board[1][1] = "O"
	board[2][0] = "X"

	for _, row := range board {
		fmt.Println(" ", row)
	}

	var nums [2][3]int
	n := 1
	for i := range nums {
		for j := range nums[i] {
			nums[i][j] = n
			n++
		}
	}

	for _, row := range nums {
		row[0] = 999
	}
	fmt.Println("write to range copy:", nums)

	for i := range nums {
		nums[i][0] = 999
	}
	fmt.Println("write through index:", nums)
	fmt.Println()
}
