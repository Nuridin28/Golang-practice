package main

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

func StringBasics() {
	s := "Hello, 世界"
	fmt.Println("string:", s)
	fmt.Println("len() (bytes):", len(s))
	fmt.Println("utf8.RuneCountInString (characters):", utf8.RuneCountInString(s))
	fmt.Println()
}

func IterationDemo() {
	s := "hêllo"

	fmt.Println("incorrect: indexing s[i] yields a byte, not a rune")
	for i := 0; i < len(s); i++ {
		fmt.Printf("  position %d: %c\n", i, s[i])
	}

	fmt.Println("correct: range decodes each rune")
	for i, r := range s {
		fmt.Printf("  position %d: %c\n", i, r)
	}
	fmt.Println()
}

func SecondRuneDemo() {
	s := "hêllo"

	runes := []rune(s)
	fmt.Println("via []rune(s)[1]:", string(runes[1]))

	_, size := utf8.DecodeRuneInString(s)
	second, _ := utf8.DecodeRuneInString(s[size:])
	fmt.Println("via utf8.DecodeRuneInString:", string(second))
	fmt.Println()
}

func TrimDemo() {
	s := "123oxo"

	fmt.Printf("TrimRight(%q, %q)  = %q\n", s, "xo", strings.TrimRight(s, "xo"))
	fmt.Printf("TrimSuffix(%q, %q) = %q\n", s, "xo", strings.TrimSuffix(s, "xo"))

	p := "xoxo123"
	fmt.Printf("TrimLeft(%q, %q)   = %q\n", p, "xo", strings.TrimLeft(p, "xo"))
	fmt.Printf("TrimPrefix(%q, %q) = %q\n", p, "xo", strings.TrimPrefix(p, "xo"))
	fmt.Println()
}

func BuilderGrowDemo() {
	values := []string{"Go", " ", "is", " ", "fun", "!"}

	totalLength := 0
	for _, v := range values {
		totalLength += len(v)
	}

	var sb strings.Builder
	sb.Grow(totalLength)
	for _, v := range values {
		sb.WriteString(v)
	}

	fmt.Println("concatenated:", sb.String())
	fmt.Println()
}
