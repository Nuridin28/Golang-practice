// Task 5: Find the leak. A small subslice or substring keeps the whole
// big array alive.
//
//	go run ./task5
package main

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
)

const n = 100 // 100 "files" of 1 MB

func readFile(i int) []byte {
	data := make([]byte, 1<<20) // pretend we read a 1 MB file
	copy(data, fmt.Sprintf("HEADER-%04d", i))
	return data
}

func keepHeadersSlice(n int) [][]byte {
	var headers [][]byte
	for i := 0; i < n; i++ {
		data := readFile(i)
		headers = append(headers, data[:11]) // 11 bytes... or 1 MB?
	}
	return headers
}

func keepHeadersCopy(n int) [][]byte {
	var headers [][]byte
	for i := 0; i < n; i++ {
		data := readFile(i)
		headers = append(headers, bytes.Clone(data[:11]))
	}
	return headers
}

func readBigString(i int) string {
	return fmt.Sprintf("%05d", i) + strings.Repeat("x", 1<<20) // a 1 MB string
}

func keepPrefixesSub(n int) []string {
	var out []string
	for i := 0; i < n; i++ {
		big := readBigString(i)
		out = append(out, big[:5])
	}
	return out
}

func keepPrefixesClone(n int) []string {
	var out []string
	for i := 0; i < n; i++ {
		big := readBigString(i)
		out = append(out, strings.Clone(big[:5]))
	}
	return out
}

// liveHeapMB runs a full GC and returns the live heap in MB.
func liveHeapMB() uint64 {
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc >> 20
}

func main() {
	base := liveHeapMB()

	h1 := keepHeadersSlice(n)
	fmt.Printf("subslices:  %3d MB\n", liveHeapMB()-base)
	runtime.KeepAlive(h1) // h1 is used after the measurement: the GC can't free it before
	h1 = nil

	h2 := keepHeadersCopy(n)
	fmt.Printf("copies:     %3d MB\n", liveHeapMB()-base)
	runtime.KeepAlive(h2)
	h2 = nil

	s1 := keepPrefixesSub(n)
	fmt.Printf("substrings: %3d MB\n", liveHeapMB()-base)
	runtime.KeepAlive(s1)
	s1 = nil

	s2 := keepPrefixesClone(n)
	fmt.Printf("clones:     %3d MB\n", liveHeapMB()-base)
	runtime.KeepAlive(s2)

	// Step 3: without KeepAlive the slice is dead after the call, and the
	// GC is allowed to free it before we measure.
	noKeep := keepHeadersSlice(n)
	_ = noKeep
	fmt.Printf("subslices without KeepAlive: %3d MB\n", liveHeapMB()-base)

	fmt.Printf("first headers: %s %s\n", keepHeadersSlice(1)[0], keepPrefixesClone(1)[0])
}
