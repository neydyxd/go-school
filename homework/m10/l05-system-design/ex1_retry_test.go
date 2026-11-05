package l05systemdesign

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

var errFlaky = errors.New("503 service unavailable")

// recorder подставляет Sleep и запоминает паузы.
type recorder struct {
	delays []time.Duration
	// cancelAfter: после стольких пауз отменить контекст (0: не отменять).
	cancelAfter int
	cancel      context.CancelFunc
}

func (r *recorder) sleep(ctx context.Context, d time.Duration) error {
	r.delays = append(r.delays, d)
	if r.cancelAfter > 0 && len(r.delays) == r.cancelAfter {
		r.cancel()
		return ctx.Err()
	}
	return nil
}

// failTimes возвращает fn, которая падает n раз, потом успешна, и счётчик вызовов.
func failTimes(n int, err error) (func(context.Context) error, *int) {
	calls := 0
	return func(context.Context) error {
		calls++
		if calls <= n {
			return fmt.Errorf("попытка %d: %w", calls, err)
		}
		return nil
	}, &calls
}

func ms(n int) time.Duration { return time.Duration(n) * time.Millisecond }

func TestRetryBackoff(t *testing.T) {
	tests := []struct {
		name       string
		policy     RetryPolicy
		fails      int
		wantCalls  int
		wantDelays []time.Duration
		wantErr    bool
	}{
		{"успех с первого раза", RetryPolicy{Attempts: 5, Base: ms(100), Max: ms(1000)}, 0, 1, nil, false},
		{"успех с третьего", RetryPolicy{Attempts: 5, Base: ms(100), Max: ms(1000)}, 2, 3, []time.Duration{ms(100), ms(200)}, false},
		{"потолок Max", RetryPolicy{Attempts: 6, Base: ms(100), Max: ms(500)}, 10, 6,
			[]time.Duration{ms(100), ms(200), ms(400), ms(500), ms(500)}, true},
		{"после последней попытки не спим", RetryPolicy{Attempts: 3, Base: ms(10), Max: ms(1000)}, 10, 3,
			[]time.Duration{ms(10), ms(20)}, true},
		{"Attempts 0 значит одна попытка", RetryPolicy{Attempts: 0, Base: ms(10), Max: ms(100)}, 10, 1, nil, true},
		{"full jitter применяется к уже ограниченной паузе", RetryPolicy{Attempts: 5, Base: ms(100), Max: ms(300),
			Jitter: func(d time.Duration) time.Duration { return d / 2 }}, 10, 5,
			[]time.Duration{ms(50), ms(100), ms(150), ms(150)}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &recorder{}
			tt.policy.Sleep = r.sleep
			fn, calls := failTimes(tt.fails, errFlaky)
			err := Retry(context.Background(), tt.policy, fn)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Retry error = %v, wantErr %v", err, tt.wantErr)
			}
			if *calls != tt.wantCalls {
				t.Errorf("fn вызвана %d раз, want %d", *calls, tt.wantCalls)
			}
			if !reflect.DeepEqual(r.delays, tt.wantDelays) {
				t.Errorf("паузы = %v, want %v", r.delays, tt.wantDelays)
			}
			if tt.wantErr {
				if !errors.Is(err, errFlaky) {
					t.Errorf("Retry error = %v, want обёртку над последней ошибкой fn", err)
				}
				want := fmt.Sprintf("после %d попыток", tt.wantCalls)
				if !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), fmt.Sprintf("попытка %d", tt.wantCalls)) {
					t.Errorf("Retry error = %q, want %q и текст последней ошибки", err, want)
				}
			}
		})
	}
}

func TestRetryNoOverflow(t *testing.T) {
	r := &recorder{}
	p := RetryPolicy{Attempts: 100, Base: ms(100), Max: 10 * time.Second, Sleep: r.sleep}
	fn, _ := failTimes(1000, errFlaky)
	Retry(context.Background(), p, fn)
	if len(r.delays) != 99 {
		t.Fatalf("пауз = %d, want 99", len(r.delays))
	}
	for i, d := range r.delays {
		if d <= 0 || d > p.Max {
			t.Fatalf("пауза #%d = %v, want в (0, %v]: переполнение Base * 2^i", i, d, p.Max)
		}
	}
	if last := r.delays[len(r.delays)-1]; last != p.Max {
		t.Errorf("последняя пауза = %v, want %v", last, p.Max)
	}
}

func TestRetryPermanent(t *testing.T) {
	r := &recorder{}
	calls := 0
	perm := fmt.Errorf("404 not found: %w", ErrPermanent)
	err := Retry(context.Background(), RetryPolicy{Attempts: 5, Base: ms(10), Max: ms(100), Sleep: r.sleep},
		func(context.Context) error { calls++; return perm })
	if err != perm { //nolint:errorlint // нужна та же ошибка без обёртки
		t.Errorf("Retry(постоянная ошибка) = %v, want ту же ошибку без обёртки", err)
	}
	if calls != 1 || len(r.delays) != 0 {
		t.Errorf("вызовов %d, пауз %d, want 1 и 0", calls, len(r.delays))
	}
}

func TestRetryContextCanceled(t *testing.T) {
	t.Run("во время паузы", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		r := &recorder{cancelAfter: 2, cancel: cancel}
		fn, calls := failTimes(100, errFlaky)
		err := Retry(ctx, RetryPolicy{Attempts: 10, Base: ms(10), Max: ms(100), Sleep: r.sleep}, fn)
		if !errors.Is(err, context.Canceled) || !errors.Is(err, errFlaky) {
			t.Errorf("Retry = %v, want и context.Canceled, и последнюю ошибку fn", err)
		}
		if *calls != 2 {
			t.Errorf("fn вызвана %d раз, want 2: после отмены новых попыток нет", *calls)
		}
	})
	t.Run("отменён до начала", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		fn, calls := failTimes(0, nil)
		err := Retry(ctx, RetryPolicy{Attempts: 3, Base: ms(10), Max: ms(100), Sleep: (&recorder{}).sleep}, fn)
		if !errors.Is(err, context.Canceled) || *calls != 0 {
			t.Errorf("Retry(отменённый ctx) = %v, вызовов %d, want context.Canceled и 0", err, *calls)
		}
	})
}

func TestRetryDefaultSleep(t *testing.T) {
	// Без подставного Sleep: настоящий таймер, но короткий, и отмена работает.
	fn, calls := failTimes(2, errFlaky)
	start := time.Now()
	err := Retry(context.Background(), RetryPolicy{Attempts: 3, Base: ms(5), Max: ms(20)}, fn)
	if err != nil || *calls != 3 {
		t.Fatalf("Retry = %v, вызовов %d, want nil и 3", err, *calls)
	}
	if d := time.Since(start); d < ms(15) {
		t.Errorf("Retry занял %v, want не меньше 15ms (паузы 5 и 10ms)", d)
	}

	ctx, cancel := context.WithTimeout(context.Background(), ms(50))
	defer cancel()
	start = time.Now()
	fn, _ = failTimes(100, errFlaky)
	done := make(chan error, 1)
	go func() { done <- Retry(ctx, RetryPolicy{Attempts: 5, Base: time.Hour, Max: time.Hour}, fn) }()
	select {
	case err = <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Retry с часовой паузой не вернулась через 3s после таймаута ctx 50ms: сон должен прерываться отменой")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Retry с часовой паузой и таймаутом 50ms = %v, want DeadlineExceeded", err)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("Retry ждал %v: сон должен прерываться отменой ctx", d)
	}
}
