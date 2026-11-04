package l06patterns

import (
	"context"
	"testing"
	"testing/synctest"
	"time"
)

// recv читает выход и запоминает, когда пришло каждое значение.
func recv(out <-chan int, start time.Time) (vals []int, at []time.Duration) {
	for v := range out {
		vals = append(vals, v)
		at = append(at, time.Since(start))
	}
	return vals, at
}

func TestThrottle(t *testing.T) {
	every := 100 * time.Millisecond

	t.Run("пачка значений растягивается", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			start := time.Now()
			vals, at := recv(Throttle(context.Background(), source(1, 2, 3, 4, 5), every), start)
			for i := range 5 {
				if i >= len(vals) || vals[i] != i+1 || at[i] != time.Duration(i)*every {
					t.Fatalf("получено %v в моменты %v, want 1..5 в 0, 100ms, 200ms, 300ms, 400ms", vals, at)
				}
			}
		})
	})

	t.Run("редкие значения без задержки", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			in := make(chan int)
			go func() {
				for i := 1; i <= 3; i++ {
					in <- i
					time.Sleep(3 * every)
				}
				close(in)
			}()
			start := time.Now()
			vals, at := recv(Throttle(context.Background(), in, every), start)
			want := []time.Duration{0, 3 * every, 6 * every}
			if len(at) != 3 || at[0] != want[0] || at[1] != want[1] || at[2] != want[2] {
				t.Fatalf("получено %v в моменты %v, want в %v: если every прошло, ждать не нужно", vals, at, want)
			}
		})
	})

	t.Run("интервал отсчитывается от последней отправки", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			in := make(chan int)
			go func() {
				in <- 1
				time.Sleep(5 * every / 2)
				in <- 2
				in <- 3 // сразу за вторым
				close(in)
			}()
			start := time.Now()
			vals, at := recv(Throttle(context.Background(), in, every), start)
			want := []time.Duration{0, 5 * every / 2, 7 * every / 2}
			if len(at) != 3 || at[0] != want[0] || at[1] != want[1] || at[2] != want[2] {
				t.Fatalf("получено %v в моменты %v, want в %v: между отправками не меньше every", vals, at, want)
			}
		})
	})

	t.Run("отмена закрывает выход", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			in := make(chan int) // никто не закроет
			out := Throttle(ctx, in, every)
			in <- 1
			<-out
			time.AfterFunc(time.Second, cancel)
			start := time.Now()
			if _, ok := <-out; ok || time.Since(start) != time.Second {
				t.Fatalf("после отмены выход должен закрыться (ok=%v через %v)", ok, time.Since(start))
			}
		})
	})
}
