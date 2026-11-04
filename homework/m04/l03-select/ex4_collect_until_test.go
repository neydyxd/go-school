package l03select

import (
	"slices"
	"testing"
	"time"
)

// ticking шлёт 0, 1, 2... каждые every, пока не закроют stop.
func ticking(every time.Duration, stop <-chan struct{}) <-chan int {
	ch := make(chan int)
	go func() {
		for i := 0; ; i++ {
			select {
			case ch <- i:
			case <-stop:
				return
			}
			select {
			case <-time.After(every):
			case <-stop:
				return
			}
		}
	}()
	return ch
}

// Подтесты идут в пузыре synctest: время фейковое, дедлайны проверяются точно.
func TestCollectUntil(t *testing.T) {
	t.Run("канал закрыт раньше дедлайна", func(t *testing.T) {
		bubble(t, func(t *testing.T) {
			var got []int
			start := time.Now()
			within(t, 2*time.Hour, "CollectUntil", func() { got = CollectUntil(source(1, 2, 3), 0, time.Hour) })
			if !slices.Equal(got, []int{1, 2, 3}) {
				t.Errorf("CollectUntil = %v, want [1 2 3]", got)
			}
			if e := time.Since(start); e != 0 {
				t.Errorf("CollectUntil вернулся через %v, хотя канал закрыт сразу: ждать дедлайн не нужно", e)
			}
		})
	})

	t.Run("лимит достигнут", func(t *testing.T) {
		bubble(t, func(t *testing.T) {
			stop := make(chan struct{})
			defer close(stop)
			var got []int
			within(t, 2*time.Hour, "CollectUntil", func() { got = CollectUntil(ticking(time.Millisecond, stop), 3, time.Hour) })
			if !slices.Equal(got, []int{0, 1, 2}) {
				t.Errorf("CollectUntil(limit=3) = %v, want [0 1 2]", got)
			}
		})
	})

	t.Run("тишина до дедлайна", func(t *testing.T) {
		bubble(t, func(t *testing.T) {
			var got []int
			start := time.Now()
			within(t, time.Second, "CollectUntil", func() { got = CollectUntil(make(chan int), 0, 50*time.Millisecond) })
			if len(got) != 0 {
				t.Errorf("CollectUntil = %v, want []", got)
			}
			if e := time.Since(start); e != 50*time.Millisecond {
				t.Errorf("CollectUntil вернулся через %v, want ровно 50ms", e)
			}
		})
	})

	// Главный тест: значения идут каждые 5ms, дедлайн 100ms. Если создавать
	// time.After на каждой итерации цикла, дедлайн сбрасывается при каждом
	// значении и функция не вернётся никогда.
	t.Run("дедлайн общий, а не между значениями", func(t *testing.T) {
		bubble(t, func(t *testing.T) {
			stop := make(chan struct{})
			defer close(stop)
			var got []int
			start := time.Now()
			within(t, 2*time.Second, "CollectUntil (таймер пересоздаётся в цикле?)", func() {
				got = CollectUntil(ticking(5*time.Millisecond, stop), 0, 100*time.Millisecond)
			})
			if e := time.Since(start); e != 100*time.Millisecond {
				t.Errorf("CollectUntil вернулся через %v, want ровно 100ms", e)
			}
			if len(got) == 0 {
				t.Fatal("CollectUntil ничего не собрал, хотя значения шли")
			}
			for i, v := range got {
				if v != i {
					t.Fatalf("CollectUntil = %v: значения потеряны или переставлены", got)
				}
			}
		})
	})
}
