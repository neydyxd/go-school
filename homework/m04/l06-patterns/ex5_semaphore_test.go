package l06patterns

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

func TestSemaphoreLimit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := NewSemaphore(3)
		var g gauge
		var wg sync.WaitGroup
		start := time.Now()
		for range 7 {
			wg.Go(func() {
				if err := s.Acquire(context.Background()); err != nil {
					t.Errorf("Acquire = %v, want nil", err)
					return
				}
				defer s.Release()
				g.enter()
				defer g.leave()
				time.Sleep(time.Second)
			})
		}
		wg.Wait()
		if m := g.max.Load(); m != 3 {
			t.Errorf("одновременно держали слот %d горутин, want 3", m)
		}
		if d := time.Since(start); d != sec(3) {
			t.Errorf("7 задач по 1s на 3 слота заняли %v, want 3s", d)
		}
	})
}

func TestSemaphoreTryAcquire(t *testing.T) {
	s := NewSemaphore(2)
	got := []bool{s.TryAcquire(), s.TryAcquire(), s.TryAcquire()}
	if got[0] != true || got[1] != true || got[2] != false {
		t.Fatalf("три TryAcquire на 2 слота = %v, want [true true false]", got)
	}
	s.Release()
	if !s.TryAcquire() {
		t.Error("после Release слот должен освободиться")
	}
}

func TestSemaphoreZero(t *testing.T) {
	s := NewSemaphore(0)
	if !s.TryAcquire() || s.TryAcquire() {
		t.Error("NewSemaphore(0) должен вести себя как семафор на 1 слот")
	}
}

func TestSemaphoreAcquireTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		s := NewSemaphore(1)
		if err := s.Acquire(context.Background()); err != nil {
			t.Fatalf("Acquire = %v", err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		start := time.Now()
		err := s.Acquire(ctx)
		if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) != 2*time.Second {
			t.Fatalf("Acquire на занятом семафоре = %v через %v, want DeadlineExceeded через 2s", err, time.Since(start))
		}
		s.Release()
		if !s.TryAcquire() {
			t.Error("неудачный Acquire не должен занимать слот")
		}
	})
}

func TestSemaphoreReleasePanics(t *testing.T) {
	defer func() {
		if r := recover(); r != "semaphore: release without acquire" {
			t.Errorf("Release без Acquire: recover() = %v, want паника \"semaphore: release without acquire\"", r)
		}
	}()
	NewSemaphore(1).Release()
}
