package main

import "testing"

var benchPayload = []byte("  Hello\r\nWorld, this is a test payload\r\n  ")

func BenchmarkProcessPayload(b *testing.B) {
	b.Run("Original", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = processPayload(benchPayload)
		}
	})

	b.Run("Optimized", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = processPayloadOptimized(benchPayload)
		}
	})
}
