package l07problems

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"
	"time"
)

// stubborn: реплика, которая игнорирует отмену и отвечает через d.
func stubborn(d time.Duration, val string, err error) Replica {
	return func(context.Context, string) (string, error) {
		time.Sleep(d)
		return val, err
	}
}

// polite: реплика, которая уважает отмену.
func polite(d time.Duration, val string) Replica {
	return func(ctx context.Context, _ string) (string, error) {
		select {
		case <-time.After(d):
			return val, nil
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}

func TestSearch(t *testing.T) {
	errA, errB := errors.New("реплика A упала"), errors.New("реплика B упала")

	t.Run("первый успех и без ожидания остальных", func(t *testing.T) {
		// synctest дождётся всех горутин в конце. Если горутина реплики
		// повиснет на отправке результата, тест упадёт с deadlock: это утечка.
		// Время в пузыре фейковое, поэтому «час» проходит мгновенно.
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			got, err := Search(context.Background(), "go",
				stubborn(time.Hour, "медленная", nil),
				stubborn(time.Second, "быстрая", nil),
				stubborn(time.Minute, "", errA),
			)
			if got != "быстрая" || err != nil {
				t.Fatalf("Search = %q, %v, want %q, nil", got, err, "быстрая")
			}
			if d := time.Since(start); d != time.Second {
				t.Errorf("Search длился %v, want 1s: остальных ждать не нужно", d)
			}
			// Даём упрямым репликам доработать. Их горутины обязаны завершиться;
			// застрявшая на отправке результата горутина и есть утечка.
			time.Sleep(2 * time.Hour)
		})
	})

	t.Run("остальным отменяют контекст", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			canceled := make(chan struct{})
			watcher := func(ctx context.Context, _ string) (string, error) {
				<-ctx.Done()
				close(canceled)
				return "", ctx.Err()
			}
			if _, err := Search(context.Background(), "go", polite(time.Second, "ok"), watcher); err != nil {
				t.Fatalf("Search = %v, want nil", err)
			}
			select {
			case <-canceled:
			case <-time.After(time.Hour):
				t.Error("проигравшей реплике не отменили контекст")
			}
		})
	})

	t.Run("ошибка не мешает успеху", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			got, err := Search(context.Background(), "go", stubborn(0, "", errA), polite(time.Second, "ok"))
			if got != "ok" || err != nil {
				t.Fatalf("Search = %q, %v, want %q, nil", got, err, "ok")
			}
		})
	})

	t.Run("все упали", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			_, err := Search(context.Background(), "go", stubborn(time.Second, "", errA), stubborn(2*time.Second, "", errB))
			if !errors.Is(err, ErrAllFailed) || !errors.Is(err, errA) || !errors.Is(err, errB) {
				t.Fatalf("Search = %v, want ErrAllFailed и обе ошибки внутри", err)
			}
		})
	})

	t.Run("без реплик", func(t *testing.T) {
		if _, err := Search(context.Background(), "go"); !errors.Is(err, ErrAllFailed) {
			t.Fatalf("Search() = %v, want ErrAllFailed", err)
		}
	})

	t.Run("родителя отменили", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			start := time.Now()
			_, err := Search(ctx, "go", stubborn(time.Hour, "поздно", nil))
			if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) != time.Second {
				t.Fatalf("Search = %v через %v, want DeadlineExceeded через 1s", err, time.Since(start))
			}
			time.Sleep(2 * time.Hour) // реплика ответит, её горутина должна завершиться
		})
	})
}
