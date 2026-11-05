package l01whyqueues

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

var errTemp = errors.New("временная ошибка")

// runRetry запускает Retry в горутине и валит тест, если она не вернулась за 2 секунды.
func runRetry(t *testing.T, ctx context.Context, max int, wait func(int) time.Duration, fn func(context.Context) error) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- Retry(ctx, max, wait, fn) }()
	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		t.Fatal("Retry не вернулась за 2 секунды: ждёт после последней попытки или не слушает ctx")
		return nil
	}
}

func TestRetry(t *testing.T) {
	tests := []struct {
		name      string
		max       int
		failFirst int   // сколько первых вызовов fn вернут ошибку
		failErr   error // какую
		wantCalls int
		wantWaits []int
		wantErr   error
	}{
		{"успех с первой попытки", 3, 0, errTemp, 1, nil, nil},
		{"успех с третьей", 5, 2, errTemp, 3, []int{1, 2}, nil},
		{"попытки кончились", 3, 10, errTemp, 3, []int{1, 2}, errTemp},
		{"maxAttempts 1", 1, 10, errTemp, 1, nil, errTemp},
		{"maxAttempts 0 это одна попытка", 0, 10, errTemp, 1, nil, errTemp},
		{"permanent не повторяется", 5, 10, Permanent(errTemp), 1, nil, errTemp},
		{"обёрнутая permanent тоже", 5, 10, fmt.Errorf("шлюз: %w", Permanent(errTemp)), 1, nil, errTemp},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls int
			var waits []int
			err := runRetry(t, context.Background(), tt.max,
				func(n int) time.Duration {
					waits = append(waits, n)
					if n >= tt.wantCalls { // ожидание после последней попытки: тест упадёт по таймауту
						return time.Hour
					}
					return time.Millisecond
				},
				func(context.Context) error {
					calls++
					if calls <= tt.failFirst {
						return tt.failErr
					}
					return nil
				})
			if !errors.Is(err, tt.wantErr) || (tt.wantErr == nil) != (err == nil) {
				t.Errorf("Retry(max=%d) error = %v, want %v", tt.max, err, tt.wantErr)
			}
			if calls != tt.wantCalls {
				t.Errorf("Retry(max=%d) вызвала fn %d раз, want %d", tt.max, calls, tt.wantCalls)
			}
			if !reflect.DeepEqual(waits, tt.wantWaits) {
				t.Errorf("Retry(max=%d) вызывала wait с %v, want %v", tt.max, waits, tt.wantWaits)
			}
		})
	}
}

func TestRetryPermanentKeepsType(t *testing.T) {
	err := runRetry(t, context.Background(), 3, func(int) time.Duration { return time.Millisecond },
		func(context.Context) error { return Permanent(errTemp) })
	if _, ok := errors.AsType[*PermanentError](err); !ok {
		t.Errorf("Retry вернула %v, want ошибку, из которой errors.As достаёт *PermanentError", err)
	}
}

func TestRetryCancelDuringWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var calls int
	err := runRetry(t, ctx, 5, func(int) time.Duration { return time.Hour },
		func(context.Context) error {
			calls++
			cancel() // отмена прилетает, пока Retry собирается ждать час
			return errTemp
		})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Retry после отмены во время ожидания = %v, want context.Canceled", err)
	}
	if calls != 1 {
		t.Errorf("fn вызвана %d раз, want 1", calls)
	}
}

func TestRetryCanceledBeforeStart(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	var calls int
	err := runRetry(t, ctx, 3, func(int) time.Duration { return time.Millisecond },
		func(context.Context) error { calls++; return nil })
	if !errors.Is(err, context.Canceled) || calls != 0 {
		t.Errorf("Retry с отменённым ctx = (%v, fn вызвана %d раз), want (context.Canceled, 0)", err, calls)
	}
}
