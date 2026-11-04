package l05context

import (
	"context"
	"errors"
	"runtime"
	"testing"
	"testing/synctest"
	"time"
)

type tagKey struct{}

func TestMerge(t *testing.T) {
	errA := errors.New("причина A")
	errB := errors.New("причина B")

	t.Run("отменили b", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			a, cancelA := context.WithCancelCause(context.WithValue(context.Background(), tagKey{}, "из a"))
			defer cancelA(nil)
			b, cancelB := context.WithCancelCause(context.Background())
			ctx, cancel := Merge(a, b)
			defer cancel()

			if got, _ := ctx.Value(tagKey{}).(string); got != "из a" {
				t.Errorf("ctx.Value(tagKey) = %q, want %q", got, "из a")
			}
			synctest.Wait()
			if ctx.Err() != nil {
				t.Fatalf("до отмены ctx.Err() = %v, want nil", ctx.Err())
			}
			cancelB(errB)
			synctest.Wait()
			if !errors.Is(ctx.Err(), context.Canceled) || !errors.Is(context.Cause(ctx), errB) {
				t.Errorf("после отмены b: Err = %v, Cause = %v, want Canceled, %v", ctx.Err(), context.Cause(ctx), errB)
			}
		})
	})

	t.Run("отменили a", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			a, cancelA := context.WithCancelCause(context.Background())
			b, cancelB := context.WithCancel(context.Background())
			defer cancelB()
			ctx, cancel := Merge(a, b)
			defer cancel()
			cancelA(errA)
			synctest.Wait()
			if !errors.Is(context.Cause(ctx), errA) {
				t.Errorf("после отмены a: Cause = %v, want %v", context.Cause(ctx), errA)
			}
		})
	})

	t.Run("дедлайн берётся из a", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			a, cancelA := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancelA()
			b, cancelB := context.WithTimeout(context.Background(), time.Hour)
			defer cancelB()
			start := time.Now()
			ctx, cancel := Merge(a, b)
			defer cancel()
			if dl, ok := ctx.Deadline(); !ok || dl.Sub(start) != 3*time.Second {
				t.Errorf("ctx.Deadline() = start+%v, %v, want start+3s, true", dl.Sub(start), ok)
			}
			<-ctx.Done()
			if !errors.Is(ctx.Err(), context.DeadlineExceeded) || time.Since(start) != 3*time.Second {
				t.Errorf("ctx закрылся через %v с %v, want 3s и DeadlineExceeded", time.Since(start), ctx.Err())
			}
		})
	})

	t.Run("cancel не оставляет горутин", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			b, cancelB := context.WithCancel(context.Background())
			defer cancelB()
			synctest.Wait()
			before := runtime.NumGoroutine()
			const n = 20
			for range n {
				ctx, cancel := Merge(context.Background(), b)
				cancel()
				<-ctx.Done()
			}
			synctest.Wait()
			if extra := runtime.NumGoroutine() - before; extra >= n/2 {
				t.Errorf("после %d вызовов Merge+cancel осталось %d лишних горутин, want 0: b ещё жив, а горутина его ждёт", n, extra)
			}
		})
	})
}
