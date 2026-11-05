package l07testingservice

import (
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// runWithin запускает f и валит тест, если она не вернулась за limit.
func runWithin(t *testing.T, limit time.Duration, f func()) time.Duration {
	t.Helper()
	done := make(chan time.Duration, 1)
	go func() {
		start := time.Now()
		f()
		done <- time.Since(start)
	}()
	select {
	case d := <-done:
		return d
	case <-time.After(limit):
		t.Fatalf("Eventually не вернулся за %v", limit)
		return 0
	}
}

func TestEventuallySuccess(t *testing.T) {
	var calls atomic.Int32
	rec := &recorder{}
	runWithin(t, 2*time.Second, func() {
		Eventually(rec, func() bool { return calls.Add(1) == 3 }, time.Second, 5*time.Millisecond)
	})
	if len(rec.errors) != 0 {
		t.Errorf("условие выполнилось, но Errorf вызван: %v", rec.errors)
	}
	if n := calls.Load(); n != 3 {
		t.Errorf("cond вызван %d раз, want 3: после true проверять больше не нужно", n)
	}
}

func TestEventuallyImmediate(t *testing.T) {
	rec := &recorder{}
	d := runWithin(t, 2*time.Second, func() {
		Eventually(rec, func() bool { return true }, time.Hour, time.Hour)
	})
	if d > 500*time.Millisecond || len(rec.errors) != 0 {
		t.Errorf("cond сразу true: ждали %v, ошибки %v; want мгновенный выход без ошибок", d, rec.errors)
	}
}

func TestEventuallyTimeout(t *testing.T) {
	var calls atomic.Int32
	rec := &recorder{}
	d := runWithin(t, 3*time.Second, func() {
		Eventually(rec, func() bool { calls.Add(1); return false }, 100*time.Millisecond, 10*time.Millisecond)
	})
	if d < 100*time.Millisecond {
		t.Errorf("вышел через %v, раньше таймаута 100ms", d)
	}
	if len(rec.errors) != 1 {
		t.Fatalf("Errorf вызван %d раз, want ровно 1: %v", len(rec.errors), rec.errors)
	}
	if !strings.Contains(rec.errors[0], "100ms") {
		t.Errorf("сообщение %q не содержит таймаут 100ms", rec.errors[0])
	}
	if n := calls.Load(); n < 2 {
		t.Errorf("за 100ms с тиком 10ms cond вызван %d раз: want несколько проверок", n)
	}
	after := calls.Load()
	time.Sleep(50 * time.Millisecond) // проверяем, что фоновых проверок не осталось
	if calls.Load() != after {
		t.Errorf("cond вызывается после выхода из Eventually: утечка горутины или тикера")
	}
}
