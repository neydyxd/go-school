package l05context

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

var errBoom = errors.New("бум")
var errShutdown = errors.New("сервер останавливается")

// waitFn ждёт d, уважая контекст, и возвращает result.
func waitFn(d time.Duration, result error) func(context.Context) error {
	return func(ctx context.Context) error {
		if _, ok := ctx.Deadline(); !ok {
			return errors.New("у контекста fn нет дедлайна")
		}
		select {
		case <-time.After(d):
			return result
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func TestCall(t *testing.T) {
	tests := []struct {
		name    string
		parent  func(context.Context) (context.Context, func())
		fn      func(context.Context) error
		want    error
		notWant error
		wantDur time.Duration
	}{
		{"успела без ошибки", nil, waitFn(time.Second, nil), nil, nil, time.Second},
		{"успела с ошибкой", nil, waitFn(time.Second, errBoom), errBoom, nil, time.Second},
		{"свой таймаут", nil, waitFn(time.Minute, nil), ErrSlow, nil, 5 * time.Second},
		{"родителя отменили с причиной", func(ctx context.Context) (context.Context, func()) {
			ctx, cancel := context.WithCancelCause(ctx)
			time.AfterFunc(time.Second, func() { cancel(errShutdown) })
			return ctx, func() { cancel(nil) }
		}, waitFn(time.Minute, nil), errShutdown, ErrSlow, time.Second},
		{"у родителя дедлайн короче", func(ctx context.Context) (context.Context, func()) {
			ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
			return ctx, cancel
		}, waitFn(time.Minute, nil), context.DeadlineExceeded, ErrSlow, 2 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx := context.Background()
				if tt.parent != nil {
					var cancel func()
					ctx, cancel = tt.parent(ctx)
					defer cancel()
				}
				start := time.Now()
				err := Call(ctx, 5*time.Second, tt.fn)
				if tt.want == nil && err != nil || !errors.Is(err, tt.want) {
					t.Errorf("Call(...) = %v, want %v", err, tt.want)
				}
				if tt.notWant != nil && errors.Is(err, tt.notWant) {
					t.Errorf("Call(...) = %v, а это не та причина: %v тут ни при чём", err, tt.notWant)
				}
				if d := time.Since(start); d != tt.wantDur {
					t.Errorf("Call длился %v, want %v", d, tt.wantDur)
				}
			})
		})
	}
}
