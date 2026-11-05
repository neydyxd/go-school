package l02livecoding

import (
	"context"
	"slices"
	"strconv"
	"testing"
	"testing/synctest"
	"time"
)

func TestPipeline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		sq := Stage(ctx, Gen(ctx, 1), func(v int) int { return v * v })
		str := Stage(ctx, sq, strconv.Itoa)
		got := Take(ctx, str, 5)
		if want := []string{"1", "4", "9", "16", "25"}; !slices.Equal(got, want) {
			t.Errorf("Take(Stage(Stage(Gen(1))), 5) = %v, want %v", got, want)
		}
		cancel() // после отмены все горутины конвейера обязаны выйти, иначе synctest увидит утечку
	})
}

func TestStageClosedInput(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		in := make(chan int, 3)
		in <- 1
		in <- 2
		in <- 3
		close(in)
		var got []int
		for v := range Stage(t.Context(), in, func(v int) int { return -v }) {
			got = append(got, v)
		}
		if want := []int{-1, -2, -3}; !slices.Equal(got, want) {
			t.Errorf("Stage по закрытому входу = %v, want %v (и закрытый выход)", got, want)
		}
	})
}

func TestStageCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		in := make(chan int) // никогда не закрывается
		out := Stage(ctx, in, func(v int) int { return v })
		in <- 1 // Stage прочитал значение и пытается отправить его, а читателя нет
		cancel()
		synctest.Wait()
		select {
		case _, ok := <-out:
			if ok {
				t.Fatalf("после отмены Stage отдал значение: при выходе он должен закрыть канал, а не ждать читателя")
			}
		default:
			t.Fatalf("после отмены Stage не закрыл выходной канал")
		}
	})
}

func TestTake(t *testing.T) {
	t.Run("отмена посреди чтения", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 2500*time.Millisecond)
			defer cancel()
			start := time.Now()
			slow := Stage(ctx, Gen(ctx, 0), func(v int) int { time.Sleep(time.Second); return v })
			got := Take(ctx, slow, 10)
			if d := time.Since(start); d != 2500*time.Millisecond {
				t.Errorf("Take вернулся через %v, want 2.5s: по отмене ctx надо выходить сразу", d)
			}
			if !slices.Equal(got, []int{0, 1}) {
				t.Errorf("Take с таймаутом 2.5s и стадией 1s/значение = %v, want [0 1]", got)
			}
			time.Sleep(time.Second) // стадия досыпает текущий fn и выходит по отмене
		})
	})
	t.Run("вход закрылся раньше n", func(t *testing.T) {
		in := make(chan int, 2)
		in <- 7
		close(in)
		if got := Take(t.Context(), in, 5); !slices.Equal(got, []int{7}) {
			t.Errorf("Take(закрытый канал с одним 7, 5) = %v, want [7]", got)
		}
	})
	t.Run("n <= 0 ничего не читает", func(t *testing.T) {
		in := make(chan int, 1)
		in <- 1
		for _, n := range []int{0, -1} {
			got := Take(t.Context(), in, n)
			if got == nil || len(got) != 0 || len(in) != 1 {
				t.Errorf("Take(n=%d) = %#v, в канале осталось %d; want пустой не-nil и нетронутый канал", n, got, len(in))
			}
		}
	})
}
