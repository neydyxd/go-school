package l02livecoding

import (
	"context"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

// feed отправляет задачи в канал и закрывает его; выходит при отмене ctx.
func feed(ctx context.Context, n int) <-chan int {
	jobs := make(chan int)
	go func() {
		defer close(jobs)
		for i := range n {
			select {
			case jobs <- i:
			case <-ctx.Done():
				return
			}
		}
	}()
	return jobs
}

func TestRunPool(t *testing.T) {
	tests := []struct {
		name    string
		jobs    int
		workers int
		wantMax int32
		wantDur time.Duration
	}{
		{"10 задач, 3 воркера", 10, 3, 3, sec(4)},
		{"воркеров больше задач", 2, 5, 2, sec(1)},
		{"0 воркеров значит 1", 3, 0, 1, sec(3)},
		{"нет задач", 0, 4, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var g gauge
				start := time.Now()
				out := RunPool(t.Context(), feed(t.Context(), tt.jobs), tt.workers, func(j int) int {
					g.enter()
					defer g.leave()
					time.Sleep(time.Second)
					return j * 10
				})
				var got []int
				for r := range out { // канал обязан закрыться, иначе synctest сообщит о deadlock
					got = append(got, r)
				}
				slices.Sort(got)
				want := make([]int, tt.jobs)
				for i := range want {
					want[i] = i * 10
				}
				if !slices.Equal(got, want) {
					t.Errorf("RunPool(%d задач) вернул %v, want %v", tt.jobs, got, want)
				}
				if m := g.max.Load(); m != tt.wantMax {
					t.Errorf("одновременно шло %d вызовов handle, want %d", m, tt.wantMax)
				}
				if d := time.Since(start); d != tt.wantDur {
					t.Errorf("пул работал %v, want %v", d, tt.wantDur)
				}
			})
		})
	}
}

func TestRunPoolCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		out := RunPool(ctx, feed(ctx, 1_000_000), 3, func(j int) int {
			time.Sleep(time.Second)
			return j
		})
		<-out
		<-out
		cancel()
		time.Sleep(5 * time.Second) // текущие вызовы handle доработают
		synctest.Wait()             // результаты никто не читает: воркеры обязаны выйти по ctx
		select {
		case _, ok := <-out:
			if ok {
				t.Fatalf("после отмены из канала пришёл результат: воркер взял новую задачу или ждал на отправке вместо выхода")
			}
		default:
			t.Fatalf("после отмены и выхода воркеров канал результатов не закрыт")
		}
	})
}
