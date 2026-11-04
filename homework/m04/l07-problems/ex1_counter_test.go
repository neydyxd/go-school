package l07problems

import (
	"sync"
	"testing"
)

// Запускай с детектором гонок: go test -race ./m04/l07-problems/
// Без -race гонка в Counter может и не проявиться.

func TestCounterZeroValue(t *testing.T) {
	var c Counter
	if got := c.Get("x"); got != 0 {
		t.Errorf("Get у пустого счётчика = %d, want 0", got)
	}
	c.Inc("x")
	if got := c.Get("x"); got != 1 {
		t.Errorf("после Inc Get = %d, want 1", got)
	}
	var empty Counter
	if s := empty.Snapshot(); s == nil || len(s) != 0 {
		t.Errorf("Snapshot пустого счётчика = %#v, want пустая не-nil мапа", s)
	}
}

func TestCounterConcurrent(t *testing.T) {
	var c Counter
	var wg sync.WaitGroup
	const goroutines, perG = 50, 200
	for i := range goroutines {
		key := []string{"a", "b"}[i%2]
		wg.Go(func() {
			for range perG {
				c.Inc(key)
				_ = c.Get(key)
				_ = c.Snapshot()
			}
		})
	}
	wg.Wait()
	want := goroutines / 2 * perG
	if a, b := c.Get("a"), c.Get("b"); a != want || b != want {
		t.Errorf("после конкурентных Inc: a=%d, b=%d, want по %d", a, b, want)
	}
}

func TestCounterSnapshotIsCopy(t *testing.T) {
	var c Counter
	c.Inc("x")
	s := c.Snapshot()
	s["x"] = 100
	s["y"] = 1
	if got := c.Get("x"); got != 1 {
		t.Errorf("запись в Snapshot изменила счётчик: Get(x) = %d, want 1", got)
	}
	c.Inc("x")
	if s["x"] != 100 {
		t.Errorf("Inc изменил ранее выданный Snapshot: s[x] = %d, want 100", s["x"])
	}
}
