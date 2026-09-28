package main

import (
	"fmt"
	"os"
)

func main() {
	section := "all"
	if len(os.Args) > 1 {
		section = os.Args[1]
	}

	switch section {
	case "1", "arrays":
		runArrays()
	case "2", "slices":
		runSlices()
	case "all":
		runArrays()
		runSlices()
	default:
		fmt.Printf("unknown section %q; use 1, 2, or no argument\n", section)
		os.Exit(1)
	}
}

func runArrays() {
	Task1_1()
	Task1_2()
	Task1_3()
	ParticipationTask1()
}

func runSlices() {
	LenCapDemo()
	Task2_1()
	Task2_2()
	ParticipationTask2()
}
