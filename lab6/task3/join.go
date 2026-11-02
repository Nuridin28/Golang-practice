// Task 3: Remove the allocations.
// Run: go test -v ./task3
//
//	go test -run='^$' -bench=. -benchmem ./task3
//	go build -o /dev/null -gcflags='-m -l' ./task3
package task3

import (
	"fmt"
	"strconv"
	"strings"
)

// JoinIDsSlow returns "1,2,3": the naive way.
func JoinIDsSlow(ids []int) string {
	s := ""
	for i, id := range ids {
		if i > 0 {
			s += ","
		}
		s += fmt.Sprint(id)
	}
	return s
}

// JoinIDsFast returns the same result, using one buffer.
func JoinIDsFast(ids []int) string {
	var sb strings.Builder
	sb.Grow(len(ids) * 8)
	var tmp [20]byte
	for i, id := range ids {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.Write(strconv.AppendInt(tmp[:0], int64(id), 10))
	}
	return sb.String()
}
