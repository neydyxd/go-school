package l05context

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

// worker возвращает функцию, которая через d вернёт (val, err), но раньше
// выйдет с ctx.Err(), если контекст отменят. Счётчики фиксируют, сколько
// функций завершилось и сколько из них увидели отмену.
func worker(d time.Duration, val string, err error, finished, canceled *atomic.Int32) func(context.Context) (string, error) {
	return func(ctx context.Context) (string, error) {
		defer finished.Add(1)
		select {
		case <-time.After(d):
			return val, err
		case <-ctx.Done():
			canceled.Add(1)
			return "", ctx.Err()
		}
	}
}

func TestFirst(t *testing.T) {
	errA := errors.New("ошибка A")
	errB := errors.New("ошибка B")

	t.Run("побеждает самая быстрая, остальных отменяют", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var finished, canceled atomic.Int32
			start := time.Now()
			got, err := First(context.Background(),
				worker(3*time.Second, "медленная", nil, &finished, &canceled),
				worker(1*time.Second, "быстрая", nil, &finished, &canceled),
				worker(2*time.Second, "средняя", nil, &finished, &canceled),
			)
			if got != "быстрая" || err != nil {
				t.Fatalf("First(...) = %q, %v, want %q, nil", got, err, "быстрая")
			}
			if d := time.Since(start); d != time.Second {
				t.Errorf("First длилась %v, want 1s: остальных надо отменить, а не ждать", d)
			}
			if n := finished.Load(); n != 3 {
				t.Errorf("к возврату First завершилось %d функций из 3: First обязана дождаться всех", n)
			}
			if n := canceled.Load(); n != 2 {
				t.Errorf("отмену увидели %d функций, want 2", n)
			}
		})
	})

	t.Run("ошибка не мешает успеху", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var finished, canceled atomic.Int32
			got, err := First(context.Background(),
				worker(1*time.Second, "", errA, &finished, &canceled),
				worker(2*time.Second, "ок", nil, &finished, &canceled),
			)
			if got != "ок" || err != nil {
				t.Fatalf("First(...) = %q, %v, want %q, nil", got, err, "ок")
			}
		})
	})

	t.Run("все упали", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var finished, canceled atomic.Int32
			got, err := First(context.Background(),
				worker(1*time.Second, "", errA, &finished, &canceled),
				worker(2*time.Second, "", errB, &finished, &canceled),
			)
			if got != "" || !errors.Is(err, errA) || !errors.Is(err, errB) {
				t.Fatalf("First(...) = %q, %v, want \"\" и обе ошибки внутри", got, err)
			}
		})
	})

	t.Run("родительский контекст отменили", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var finished, canceled atomic.Int32
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			start := time.Now()
			_, err := First(ctx,
				worker(5*time.Second, "а", nil, &finished, &canceled),
				worker(6*time.Second, "б", nil, &finished, &canceled),
			)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("First(...) err = %v, want DeadlineExceeded внутри", err)
			}
			if d := time.Since(start); d != time.Second {
				t.Errorf("First длилась %v, want 1s", d)
			}
		})
	})

	t.Run("без функций", func(t *testing.T) {
		got, err := First(context.Background())
		if got != "" || !errors.Is(err, ErrNoFuncs) {
			t.Fatalf("First() = %q, %v, want \"\", ErrNoFuncs", got, err)
		}
	})
}
