package main

import (
	"context"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

// checkNoLeak runs f and fails the test if f leaves goroutines behind.
// Goroutines that are about to finish get a short time to do it.
func checkNoLeak(t *testing.T, f func()) {
	t.Helper()
	before := runtime.NumGoroutine()
	f()
	after := runtime.NumGoroutine()
	for deadline := time.Now().Add(300 * time.Millisecond); after > before && time.Now().Before(deadline); {
		time.Sleep(10 * time.Millisecond)
		after = runtime.NumGoroutine()
	}
	if after > before {
		var sb strings.Builder
		pprof.Lookup("goroutine").WriteTo(&sb, 1)
		t.Fatalf("%d goroutine(s) leaked: %d before, %d after\n%s", after-before, before, after, sb.String())
	}
}

func TestFetchFixed_NoLeak(t *testing.T) {
	checkNoLeak(t, func() {
		for i := 0; i < 10; i++ {
			fetchFixed(10 * time.Millisecond)
		}
	})
}

// TestFetchLeaky_NoLeak is the same check for fetchLeaky. It fails, so it is
// skipped unless you ask for it: SHOW_LEAK=1 go test -run Leaky -v ./task5
func TestFetchLeaky_NoLeak(t *testing.T) {
	if os.Getenv("SHOW_LEAK") == "" {
		t.Skip("fetchLeaky leaks on purpose; run with SHOW_LEAK=1 to see this test fail")
	}
	checkNoLeak(t, func() {
		for i := 0; i < 10; i++ {
			fetchLeaky(10 * time.Millisecond)
		}
	})
}

func TestFetch_Result(t *testing.T) {
	for name, fetch := range map[string]func(time.Duration) (int, error){
		"fetchLeaky": fetchLeaky,
		"fetchFixed": fetchFixed,
	} {
		if v, err := fetch(time.Second); v != 42 || err != nil {
			t.Errorf("%s(1s) = %d, %v; want 42, nil", name, v, err)
		}
		if _, err := fetch(time.Millisecond); err == nil {
			t.Errorf("%s(1ms): want a timeout error", name)
		}
	}
}

func TestStartTicker_StopsOnCancel(t *testing.T) {
	checkNoLeak(t, func() {
		ctx, cancel := context.WithCancel(context.Background())
		for i := 0; i < 10; i++ {
			startTicker(ctx)
		}
		time.Sleep(30 * time.Millisecond)
		cancel()
	})
}

// The same idea with testing/synctest (Go 1.25): the test runs in a "bubble"
// with a fake clock, so sleeping takes no real time, and synctest.Test fails
// with "deadlock" if a goroutine of the bubble is still blocked at the end.
//
// The fake clock only moves while the bubble's main goroutine is alive, so we
// wait one (fake) minute at the end: every goroutine that can finish, finishes.
func TestFetchFixed_Synctest(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		if _, err := fetchFixed(10 * time.Millisecond); err == nil {
			t.Error("want a timeout error")
		}
		time.Sleep(time.Minute)
	})
}

func TestFetchLeaky_Synctest(t *testing.T) {
	if os.Getenv("SHOW_LEAK") == "" {
		t.Skip("fetchLeaky leaks on purpose; run with SHOW_LEAK=1 to see this test fail")
	}
	synctest.Test(t, func(t *testing.T) {
		fetchLeaky(10 * time.Millisecond)
		time.Sleep(time.Minute)
	})
}
