package l02livecoding

import (
	"context"
	"errors"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

const limiterHint = "Wait не возвращается: ждёт жетон под мьютексом или жетоны не накапливаются (проверь пополнение по времени)"

// near: время совпадает с точностью до миллисекунды (решение на float64 даёт наносекундные хвосты).
func near(got, want time.Duration) bool {
	return got >= want-time.Millisecond && got <= want+time.Millisecond
}

func allowN(l *Limiter, n int) (ok int) {
	for range n {
		if l.Allow() {
			ok++
		}
	}
	return ok
}

func TestLimiterAllow(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(time.Second, 3)
		if got := allowN(l, 5); got != 3 {
			t.Fatalf("полное ведро на 3: Allow разрешил %d из 5, want 3", got)
		}
		time.Sleep(time.Second)
		if got := allowN(l, 3); got != 1 {
			t.Fatalf("через 1s после опустошения разрешено %d, want 1", got)
		}
		time.Sleep(500 * time.Millisecond)
		if l.Allow() {
			t.Fatalf("через 0.5s жетона ещё нет, а Allow = true")
		}
		time.Sleep(500 * time.Millisecond)
		if !l.Allow() {
			t.Fatalf("через 0.5s + 0.5s жетон накопился, а Allow = false: прогресс пополнения теряется между вызовами")
		}
		time.Sleep(10 * time.Second)
		if got := allowN(l, 10); got != 3 {
			t.Fatalf("после 10s простоя разрешено %d, want 3: больше burst копиться не должно", got)
		}
	})
}

func TestLimiterEdge(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(time.Second, 0)
		if got := allowN(l, 3); got != 1 {
			t.Errorf("burst 0 считается 1: разрешено %d, want 1", got)
		}
		free := NewLimiter(0, 1)
		if got := allowN(free, 100); got != 100 {
			t.Errorf("every 0 значит без ограничения: разрешено %d из 100", got)
		}
		start := time.Now()
		if err := free.Wait(t.Context()); err != nil || time.Since(start) != 0 {
			t.Errorf("Wait без ограничения = %v за %v, want nil сразу", err, time.Since(start))
		}
	})
}

func TestLimiterWait(t *testing.T) {
	defer watchdog(t, limiterHint)()
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(time.Second, 1)
		start := time.Now()
		for i := range 3 {
			if err := l.Wait(t.Context()); err != nil {
				t.Fatalf("Wait #%d = %v", i+1, err)
			}
		}
		if d := time.Since(start); !near(d, sec(2)) {
			t.Errorf("три Wait при burst 1 и 1/s заняли %v, want 2s (первый сразу)", d)
		}
	})
}

func TestLimiterWaitCancel(t *testing.T) {
	defer watchdog(t, limiterHint)()
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(time.Second, 1)
		l.Allow()
		ctx, cancel := context.WithTimeout(t.Context(), 300*time.Millisecond)
		defer cancel()
		start := time.Now()
		err := l.Wait(ctx)
		if !errors.Is(err, context.DeadlineExceeded) || !near(time.Since(start), 300*time.Millisecond) {
			t.Fatalf("Wait с таймаутом 300ms = %v через %v, want DeadlineExceeded через 300ms", err, time.Since(start))
		}
		time.Sleep(700 * time.Millisecond)
		if !l.Allow() {
			t.Fatalf("отменённый Wait потратил жетон: через 1s после опустошения Allow = false")
		}
	})
}

func TestLimiterConcurrent(t *testing.T) {
	defer watchdog(t, limiterHint)()
	synctest.Test(t, func(t *testing.T) {
		l := NewLimiter(time.Second, 2)
		start := time.Now()
		var wg sync.WaitGroup
		for range 10 {
			wg.Go(func() {
				if err := l.Wait(t.Context()); err != nil {
					t.Errorf("Wait = %v", err)
				}
			})
		}
		wg.Wait()
		if d := time.Since(start); !near(d, sec(8)) {
			t.Errorf("10 одновременных Wait при burst 2 и 1/s завершились за %v, want 8s", d)
		}
	})
}
