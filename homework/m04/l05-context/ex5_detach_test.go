package l05context

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

type userKey struct{}

func TestDetach(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		parent, cancelParent := context.WithTimeout(context.Background(), time.Second)
		defer cancelParent()
		parent = context.WithValue(parent, userKey{}, "степан")

		start := time.Now()
		ctx, cancel := Detach(parent, 10*time.Second)
		defer cancel()

		if got, _ := ctx.Value(userKey{}).(string); got != "степан" {
			t.Errorf("ctx.Value(userKey) = %q, want %q: значения родителя должны сохраниться", got, "степан")
		}
		dl, ok := ctx.Deadline()
		if !ok || dl.Sub(start) != 10*time.Second {
			t.Errorf("ctx.Deadline() = %v, %v, want start+10s, true", dl.Sub(start), ok)
		}

		cancelParent()
		time.Sleep(2 * time.Second)
		if err := ctx.Err(); err != nil {
			t.Fatalf("после отмены родителя ctx.Err() = %v, want nil", err)
		}

		time.Sleep(8 * time.Second)
		<-ctx.Done()
		if d := time.Since(start); d != 10*time.Second {
			t.Errorf("фоновый контекст закрылся через %v, want 10s", d)
		}
		if err := ctx.Err(); !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("ctx.Err() = %v, want DeadlineExceeded", err)
		}
	})

	t.Run("cancel работает", func(t *testing.T) {
		ctx, cancel := Detach(context.Background(), time.Hour)
		cancel()
		if err := ctx.Err(); !errors.Is(err, context.Canceled) {
			t.Errorf("после cancel() ctx.Err() = %v, want Canceled", err)
		}
	})
}
