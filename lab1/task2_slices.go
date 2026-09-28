package main

import "fmt"

func LenCapDemo() {
	cars := []string{"Ferrari", "Honda", "Ford", "BYD"}
	fmt.Println("cars:", cars, "has old length", len(cars), "and capacity", cap(cars))

	cars = append(cars, "Toyota")
	fmt.Println("cars:", cars, "has new length", len(cars), "and capacity", cap(cars))
	fmt.Println()
}

func Task2_1() {
	var a []int
	b := []int{1, 2, 3, 4}
	c := make([]int, 3)
	d := make([]int, 2, 10)
	arr := [6]int{10, 20, 30, 40, 50, 60}
	e := arr[1:4]

	fmt.Println("               len cap")
	fmt.Println("var a []int      ", len(a), cap(a), a)
	fmt.Println("[]int{1,2,3,4}   ", len(b), cap(b), b)
	fmt.Println("make([]int,3)    ", len(c), cap(c), c)
	fmt.Println("make([]int,2,10) ", len(d), cap(d), d)
	fmt.Println("arr[1:4]         ", len(e), cap(e), e)
	fmt.Println()
}
func Task2_2() {
	arr := [5]int{1, 2, 3, 4, 5}
	s := arr[1:3]
	s[0] = 99
	fmt.Println(arr)
	fmt.Println(s)

	a := [4]int{1, 2, 3, 4}
	b := a[0:2]
	b = append(b, 99)
	fmt.Println(a)

	c := []int{1, 2}
	d := append(c, 3)
	d[0] = 99
	fmt.Println(c)
	fmt.Println(d)

	src := []int{1, 2, 3}
	dst := make([]int, 3)
	copy(dst, src)
	src[0] = 99
	fmt.Println(src)
	fmt.Println(dst)
	fmt.Println()
}

func ParticipationTask2() {
	numbers := []int{1, 2, 3, 4, 5}
	doubleAll(numbers)
	fmt.Println("slice after doubleAll:", numbers)

	arr := [5]int{1, 2, 3, 4, 5}
	doubleArray(arr)
	fmt.Println("array after doubleArray:", arr)

	s := []int{1, 2, 3}
	appendInside(s)
	fmt.Println("after appendInside:   ", s, "len:", len(s))

	s = appendAndReturn(s)
	fmt.Println("after appendAndReturn:", s, "len:", len(s))
	fmt.Println()
}

func doubleAll(nums []int) {
	for i := range nums {
		nums[i] *= 2
	}
}

func doubleArray(nums [5]int) {
	for i := range nums {
		nums[i] *= 2
	}
}

func appendInside(nums []int) {
	nums = append(nums, 100)
}

func appendAndReturn(nums []int) []int {
	return append(nums, 100)
}
