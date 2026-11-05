package l02livecoding

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

const groupHint = "Go или Wait не возвращаются: задача не освобождает слот, Wait не дожидается задач или ожидание идёт на мьютексе (synctest его не видит, жди на канале или WaitGroup)"

func TestGroupLimit(t *testing.T) {
	defer watchdog(t, groupHint)()
	synctest.Test(t, func(t *testing.T) {
		g, _ := WithContext(t.Context())
		g.SetLimit(2)
		var gg gauge
		start := time.Now()
		for range 6 {
			g.Go(func() error {
				gg.enter()
				defer gg.leave()
				time.Sleep(time.Second)
				return nil
			})
		}
		if err := g.Wait(); err != nil {
			t.Errorf("Wait() = %v, want nil", err)
		}
		if m := gg.max.Load(); m != 2 {
			t.Errorf("одновременно работало %d задач, want 2", m)
		}
		if d := time.Since(start); d != sec(3) {
			t.Errorf("6 задач по 1s с лимитом 2 заняли %v, want 3s", d)
		}
	})
}

func TestGroupGoBlocks(t *testing.T) {
	defer watchdog(t, groupHint)()
	synctest.Test(t, func(t *testing.T) {
		var g Group
		g.SetLimit(1)
		start := time.Now()
		g.Go(func() error { time.Sleep(time.Second); return nil })
		g.Go(func() error { return nil })
		if d := time.Since(start); d != sec(1) {
			t.Errorf("второй Go при лимите 1 вернулся через %v, want 1s: Go должен ждать свободный слот", d)
		}
		g.Wait()
	})
}

func TestGroupFirstError(t *testing.T) {
	defer watchdog(t, groupHint)()
	synctest.Test(t, func(t *testing.T) {
		errA, errB := errors.New("a"), errors.New("b")
		g, ctx := WithContext(t.Context())
		var canceledAt time.Duration
		start := time.Now()
		g.Go(func() error { time.Sleep(2 * time.Second); return errB })
		g.Go(func() error { time.Sleep(time.Second); return errA })
		g.Go(func() error {
			select {
			case <-ctx.Done():
				canceledAt = time.Since(start)
			case <-time.After(time.Hour):
			}
			return nil
		})
		if err := g.Wait(); !errors.Is(err, errA) || errors.Is(err, errB) {
			t.Errorf("Wait() = %v, want первую по времени ошибку a", err)
		}
		if canceledAt != sec(1) {
			t.Errorf("контекст отменён через %v, want 1s (сразу после первой ошибки)", canceledAt)
		}
	})
}

func TestGroupWaitCancelsContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		g, ctx := WithContext(t.Context())
		g.Go(func() error { return nil })
		if err := g.Wait(); err != nil {
			t.Fatalf("Wait() = %v", err)
		}
		if !errors.Is(ctx.Err(), context.Canceled) {
			t.Errorf("после Wait ctx.Err() = %v, want context.Canceled", ctx.Err())
		}
	})
}

func TestGroupZero(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var g Group
		start := time.Now()
		for range 5 {
			g.Go(func() error { time.Sleep(time.Second); return nil })
		}
		if err := g.Wait(); err != nil || time.Since(start) != sec(1) {
			t.Errorf("нулевая Group: Wait() = %v за %v, want nil за 1s (без лимита)", err, time.Since(start))
		}
		var empty Group
		if err := empty.Wait(); err != nil {
			t.Errorf("Wait без задач = %v", err)
		}
	})
}
