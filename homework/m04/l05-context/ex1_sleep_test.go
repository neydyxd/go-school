package l05context

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

// Тесты идут внутри synctest.Test: время там фейковое и детерминированное,
// а горутина, зависшая навсегда, роняет тест вместо вечного ожидания.

func TestSleep(t *testing.T) {
	tests := []struct {
		name    string
		d       time.Duration
		setup   func(context.Context) (context.Context, context.CancelFunc)
		wantErr error
		wantDur time.Duration
	}{
		{"время вышло", 3 * time.Second, nil, nil, 3 * time.Second},
		{"нулевая длительность", 0, nil, nil, 0},
		{"отмена через секунду", 5 * time.Second, func(ctx context.Context) (context.Context, context.CancelFunc) {
			ctx, cancel := context.WithCancel(ctx)
			time.AfterFunc(time.Second, cancel)
			return ctx, cancel
		}, context.Canceled, time.Second},
		{"дедлайн раньше d", 5 * time.Second, func(ctx context.Context) (context.Context, context.CancelFunc) {
			return context.WithTimeout(ctx, 2*time.Second)
		}, context.DeadlineExceeded, 2 * time.Second},
		{"уже отменён", 5 * time.Second, func(ctx context.Context) (context.Context, context.CancelFunc) {
			ctx, cancel := context.WithCancel(ctx)
			cancel()
			return ctx, cancel
		}, context.Canceled, 0},
		{"уже отменён и d=0", 0, func(ctx context.Context) (context.Context, context.CancelFunc) {
			ctx, cancel := context.WithCancel(ctx)
			cancel()
			return ctx, cancel
		}, context.Canceled, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				ctx := context.Background()
				if tt.setup != nil {
					var cancel context.CancelFunc
					ctx, cancel = tt.setup(ctx)
					defer cancel()
				}
				start := time.Now()
				err := Sleep(ctx, tt.d)
				if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil && err != nil) {
					t.Errorf("Sleep(ctx, %v) = %v, want %v", tt.d, err, tt.wantErr)
				}
				if got := time.Since(start); got != tt.wantDur {
					t.Errorf("Sleep(ctx, %v) длился %v, want %v", tt.d, got, tt.wantDur)
				}
			})
		})
	}
}
