package main

import (
	"bytes"
	"fmt"
	"strings"
)

func processPayload(data []byte) string {
	trimmed := string(bytes.TrimSpace(data))
	clean := strings.ReplaceAll(trimmed, "\r", "")
	return clean
}

func processPayloadOptimized(data []byte) string {
	trimmed := bytes.TrimSpace(data)

	var sb strings.Builder
	sb.Grow(len(trimmed))
	for _, b := range trimmed {
		if b == '\r' {
			continue
		}
		sb.WriteByte(b)
	}
	return sb.String()
}

func PayloadDemo() {
	data := []byte("  Hello\r\nWorld\r\n  ")

	fmt.Printf("original:  %q\n", processPayload(data))
	fmt.Printf("optimized: %q\n", processPayloadOptimized(data))
	fmt.Println()
}
