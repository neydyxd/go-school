package l02livecoding

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Здесь без synctest: правильное решение обычно ждёт на мьютексе, а synctest
// не двигает фейковое время, пока кто-то ждёт мьютекс. Время настоящее, но
// короткое, и тесты проверяют счётчики, а не длительности.

func TestLazyOnce(t *testing.T) {
	defer watchdog(t, "Get не возвращается")()
	var calls atomic.Int32
	var g gauge
	l := NewLazy(func() (string, error) {
		g.enter()
		defer g.leave()
		calls.Add(1)
		time.Sleep(20 * time.Millisecond)
		return "conn", nil
	})
	var wg sync.WaitGroup
	for range 20 {
		wg.Go(func() {
			if v, err := l.Get(); v != "conn" || err != nil {
				t.Errorf("Get() = %q, %v; want conn, nil", v, err)
			}
		})
	}
	wg.Wait()
	if c := calls.Load(); c != 1 {
		t.Errorf("init вызван %d раз на 20 одновременных Get, want 1", c)
	}
	if v, err := l.Get(); v != "conn" || err != nil || calls.Load() != 1 {
		t.Errorf("Get после успеха = %q, %v, вызовов init %d; want conn, nil, 1", v, err, calls.Load())
	}
}

func TestLazyRetry(t *testing.T) {
	boom := errors.New("boom")
	calls := 0
	l := NewLazy(func() (int, error) {
		calls++
		if calls == 1 {
			return 0, boom
		}
		return 42, nil
	})
	if v, err := l.Get(); !errors.Is(err, boom) || v != 0 {
		t.Fatalf("первый Get = %v, %v; want 0, boom", v, err)
	}
	for i := range 2 {
		if v, err := l.Get(); v != 42 || err != nil {
			t.Fatalf("Get #%d после ошибки = %v, %v; want 42, nil: ошибка не должна кешироваться", i+2, v, err)
		}
	}
	if calls != 2 {
		t.Errorf("init вызван %d раз, want 2", calls)
	}
}

func TestLazyErrorsNotParallel(t *testing.T) {
	defer watchdog(t, "Get не возвращается")()
	boom := errors.New("boom")
	var g gauge
	l := NewLazy(func() (int, error) {
		g.enter()
		defer g.leave()
		time.Sleep(5 * time.Millisecond)
		return 0, boom
	})
	var wg sync.WaitGroup
	for range 8 {
		wg.Go(func() {
			if _, err := l.Get(); !errors.Is(err, boom) {
				t.Errorf("Get() = %v, want boom", err)
			}
		})
	}
	wg.Wait()
	if m := g.max.Load(); m != 1 {
		t.Errorf("одновременно шло %d вызовов init, want 1", m)
	}
}
