package l07maps

import (
	"sync"
	"testing"
	"time"
)

func TestSafeCounterBasic(t *testing.T) {
	c := NewSafeCounter()
	c.Inc("a")
	c.Inc("a")
	c.Inc("b")
	tests := []struct {
		name string
		key  string
		want int
	}{
		{"дважды", "a", 2},
		{"один раз", "b", 1},
		{"не было", "zzz", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.Get(tt.key); got != tt.want {
				t.Errorf("Get(%q) = %d, want %d", tt.key, got, tt.want)
			}
		})
	}
}

func TestSafeCounterConcurrent(t *testing.T) {
	c := NewSafeCounter()
	const workers, perWorker = 16, 1000
	var wg sync.WaitGroup
	for w := range workers {
		wg.Go(func() {
			for i := range perWorker {
				c.Inc("total")
				if i%2 == 0 {
					c.Inc("even")
				}
				_ = c.Get("total")
				if w == 0 && i%100 == 0 {
					_ = c.Snapshot()
				}
			}
		})
	}
	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("горутины не завершились за 10 секунд: похоже на дедлок (мьютекс не отпущен?)")
	}
	if got := c.Get("total"); got != workers*perWorker {
		t.Errorf("Get(\"total\") = %d, want %d", got, workers*perWorker)
	}
	if got := c.Get("even"); got != workers*perWorker/2 {
		t.Errorf("Get(\"even\") = %d, want %d", got, workers*perWorker/2)
	}
}

func TestSafeCounterSnapshotIsCopy(t *testing.T) {
	c := NewSafeCounter()
	c.Inc("x")
	snap := c.Snapshot()
	snap["x"] = 100
	snap["y"] = 1
	if got := c.Get("x"); got != 1 {
		t.Errorf("запись в Snapshot изменила счётчик: Get(\"x\") = %d, want 1", got)
	}
	c.Inc("x")
	if snap["x"] != 100 {
		t.Errorf("Inc изменил уже выданный Snapshot: snap[\"x\"] = %d, want 100", snap["x"])
	}
}
