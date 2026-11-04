package l06patterns

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

func TestFetchAll(t *testing.T) {
	t.Run("порядок и лимит", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			var g gauge
			ids := []int{5, 4, 3, 2, 1, 0}
			start := time.Now()
			got, err := FetchAll(context.Background(), ids, 2, func(ctx context.Context, id int) (string, error) {
				g.enter()
				defer g.leave()
				time.Sleep(time.Duration(id+1) * 100 * time.Millisecond) // разное время, чтобы порядок завершения перемешался
				return fmt.Sprint("id", id), nil
			})
			want := []string{"id5", "id4", "id3", "id2", "id1", "id0"}
			if err != nil || !slices.Equal(got, want) {
				t.Fatalf("FetchAll = %v, %v, want %v, nil", got, err, want)
			}
			if m := g.max.Load(); m != 2 {
				t.Errorf("одновременно шло %d вызовов, want 2", m)
			}
			if d := time.Since(start); d > 1200*time.Millisecond {
				t.Errorf("FetchAll длился %v, а с лимитом 2 хватает 1.2s: слот должен освобождаться сразу", d)
			}
		})
	})

	t.Run("первая ошибка отменяет остальных", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			errBoom := errors.New("бум")
			var calls, sawCancel atomic.Int32
			ids := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
			start := time.Now()
			got, err := FetchAll(context.Background(), ids, 2, func(ctx context.Context, id int) (string, error) {
				calls.Add(1)
				if id == 3 {
					time.Sleep(500 * time.Millisecond)
					return "", errBoom
				}
				select {
				case <-time.After(time.Second):
					return "ok", nil
				case <-ctx.Done():
					sawCancel.Add(1)
					return "", ctx.Err()
				}
			})
			if got != nil || !errors.Is(err, errBoom) {
				t.Fatalf("FetchAll = %v, %v, want nil, %v (первая ошибка, а не context.Canceled)", got, err, errBoom)
			}
			if d := time.Since(start); d != 1500*time.Millisecond {
				t.Errorf("FetchAll длился %v, want 1.5s: после ошибки остальных отменяют, а не ждут", d)
			}
			if n := calls.Load(); n > 5 {
				t.Errorf("fetch вызван %d раз, want не больше 5: после ошибки новые вызовы не начинаются", n)
			}
			if sawCancel.Load() == 0 {
				t.Error("ни один вызов не увидел отмену: контекст fetch должен быть из errgroup.WithContext")
			}
		})
	})

	t.Run("пусто", func(t *testing.T) {
		got, err := FetchAll(context.Background(), nil, 3, func(context.Context, int) (string, error) {
			return "", errors.New("не должен вызываться")
		})
		if err != nil || len(got) != 0 {
			t.Errorf("FetchAll(nil) = %v, %v, want [], nil", got, err)
		}
	})
}
