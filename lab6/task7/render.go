// Task 7: Reuse buffers with sync.Pool.
// Run: go test -v ./task7
//
//	go test -run='^$' -bench=. -benchmem ./task7
package task7

import (
	"bytes"
	"strconv"
	"sync"
)

// RenderNew makes a new 4 KB buffer on every call.
func RenderNew(ids []int) int {
	buf := new(bytes.Buffer)
	buf.Grow(4096)
	writeIDs(buf, ids)
	return buf.Len()
}

var bufPool = sync.Pool{
	New: func() any { b := new(bytes.Buffer); b.Grow(4096); return b },
}

// RenderPool takes a buffer from the pool and gives it back when done.
func RenderPool(ids []int) int {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset() // the buffer may still hold the data of the previous user
	writeIDs(buf, ids)
	n := buf.Len()
	bufPool.Put(buf) // after Put, buf belongs to the pool: don't touch it
	return n
}

// writeIDs writes "id=<n>\n" for every id, without fmt.
func writeIDs(buf *bytes.Buffer, ids []int) {
	var tmp [20]byte
	for _, id := range ids {
		buf.WriteString("id=")
		buf.Write(strconv.AppendInt(tmp[:0], int64(id), 10))
		buf.WriteByte('\n')
	}
}

// renderPoolNoReset is RenderPool without buf.Reset(), to show the bug.
func renderPoolNoReset(ids []int) int {
	buf := bufPool.Get().(*bytes.Buffer)
	writeIDs(buf, ids)
	n := buf.Len()
	bufPool.Put(buf)
	return n
}
