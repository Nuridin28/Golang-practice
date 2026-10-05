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
	case "1", "strings":
		runStrings()
	case "2", "payload":
		runPayload()
	case "3", "maps":
		runMaps()
	case "4", "inventory":
		runInventory()
	case "all":
		runStrings()
		runPayload()
		runMaps()
		runInventory()
	default:
		fmt.Printf("unknown section %q; use 1-4, strings/payload/maps/inventory, or no argument\n", section)
		os.Exit(1)
	}
}

func runStrings() {
	StringBasics()
	IterationDemo()
	SecondRuneDemo()
	TrimDemo()
	BuilderGrowDemo()
}

func runPayload() {
	PayloadDemo()
}

func runMaps() {
	MapOperationsDemo()
	ScoresTask()
}

func runInventory() {
	InventoryTask()
}
