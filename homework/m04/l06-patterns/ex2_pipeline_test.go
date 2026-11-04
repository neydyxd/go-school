package l06patterns

import (
	"context"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

func source(vals ...int) <-chan int {
	ch := make(chan int, len(vals))
	for _, v := range vals {
		ch <- v
	}
	close(ch)
	return ch
}

func TestSquare(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		got := collect(Square(context.Background(), source(1, -2, 3)))
		if want := []int{1, 4, 9}; !slices.Equal(got, want) {
			t.Errorf("Square(1, -2, 3) = %v, want %v", got, want)
		}
	})

	t.Run("отмена при незакрытом входе", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			in := make(chan int) // никто не закроет
			out := Square(ctx, in)
			in <- 2
			if v := <-out; v != 4 {
				t.Fatalf("Square отдал %d, want 4", v)
			}
			cancel()
			if _, ok := <-out; ok {
				t.Fatal("после отмены выход должен закрыться")
			}
		})
	})
}

func TestMerge(t *testing.T) {
	t.Run("собирает всё и закрывается", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			got := collect(Merge(context.Background(), source(1, 2), source(), source(3, 4, 5)))
			slices.Sort(got)
			if want := []int{1, 2, 3, 4, 5}; !slices.Equal(got, want) {
				t.Errorf("Merge(...) после сортировки = %v, want %v", got, want)
			}
		})
	})

	t.Run("без входов", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			if got := collect(Merge(context.Background())); len(got) != 0 {
				t.Errorf("Merge() = %v, want пусто и закрытый канал", got)
			}
		})
	})

	t.Run("медленный вход не тормозит быстрый", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			slow, fast := make(chan int), make(chan int)
			go func() { time.Sleep(sec(5)); slow <- 1; close(slow) }()
			go func() { fast <- 2; close(fast) }()
			start := time.Now()
			out := Merge(context.Background(), slow, fast)
			if v := <-out; v != 2 || time.Since(start) != 0 {
				t.Errorf("первым пришло %d через %v, want 2 сразу: читай все входы одновременно", v, time.Since(start))
			}
			collect(out)
		})
	})

	t.Run("отмена при незакрытых входах", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			a, b := make(chan int), make(chan int) // никто не закроет
			out := Merge(ctx, a, b)
			a <- 1
			<-out
			cancel()
			if _, ok := <-out; ok {
				t.Fatal("после отмены выход должен закрыться")
			}
		})
	})
}

func TestPipeline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := context.Background()
		got := collect(Merge(ctx, Square(ctx, Gen(ctx, 1, 2)), Square(ctx, Gen(ctx, 3))))
		slices.Sort(got)
		if want := []int{1, 4, 9}; !slices.Equal(got, want) {
			t.Errorf("Gen→Square→Merge = %v, want %v", got, want)
		}
	})
}
