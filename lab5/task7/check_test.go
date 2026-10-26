package task7

import (
	"errors"
	"testing"
)

var (
	sinkInt int
	sinkErr error
)

func TestPathsAgree(t *testing.T) {
	if _, err := checkErr(-1); err != ErrNegative {
		t.Errorf("checkErr(-1) err = %v", err)
	}
	if _, err := safeCheck(-1); err != ErrNegative {
		t.Errorf("safeCheck(-1) err = %v", err)
	}
	if _, err := checkWrapped(-1); !errors.Is(err, ErrNegative) {
		t.Errorf("checkWrapped(-1) err = %v", err)
	}
	if v, err := safeCheck(5); v != 5 || err != nil {
		t.Errorf("safeCheck(5) = %d, %v", v, err)
	}
}

func BenchmarkErrorReturn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkInt, sinkErr = checkErr(-1)
	}
}

func BenchmarkPanicRecover(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkInt, sinkErr = safeCheck(-1)
	}
}

func BenchmarkErrorWrap(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkInt, sinkErr = checkWrapped(-1)
	}
}

func BenchmarkErrorReturnOK(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkInt, sinkErr = checkErr(1)
	}
}

func BenchmarkPanicRecoverOK(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkInt, sinkErr = safeCheck(1)
	}
}

func BenchmarkIncNoDefer(b *testing.B) {
	var c Counter
	for i := 0; i < b.N; i++ {
		c.IncNoDefer()
	}
	sinkInt = c.n
}

func BenchmarkIncDefer(b *testing.B) {
	var c Counter
	for i := 0; i < b.N; i++ {
		c.IncDefer()
	}
	sinkInt = c.n
}
