package l02channels

import (
	"slices"
	"testing"
	"time"
)

// filled возвращает буферизованный канал с values; closed решает, закрыт ли он.
func filled(closed bool, values ...int) chan int {
	ch := make(chan int, len(values)+1)
	for _, v := range values {
		ch <- v
	}
	if closed {
		close(ch)
	}
	return ch
}

func TestCollect(t *testing.T) {
	tests := []struct {
		name     string
		ch       chan int
		limit    int
		want     []int
		leftover int // сколько значений должно остаться в канале
	}{
		{"закрытый пустой", filled(true), 0, nil, 0},
		{"до закрытия", filled(true, 1, 2, 3), 0, []int{1, 2, 3}, 0},
		{"отрицательный лимит = без лимита", filled(true, 1, 2), -1, []int{1, 2}, 0},
		{"лимит меньше, чем есть", filled(false, 1, 2, 3, 4), 2, []int{1, 2}, 2},
		{"лимит больше, чем есть", filled(true, 5, 6), 10, []int{5, 6}, 0},
		{"лимит ровно сколько есть, канал открыт", filled(false, 7, 8), 2, []int{7, 8}, 0},
		{"нули из буфера это значения", filled(true, 0, 0), 0, []int{0, 0}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []int
			within(t, time.Second, "Collect", func() { got = Collect(tt.ch, tt.limit) })
			if !slices.Equal(got, tt.want) {
				t.Errorf("Collect(limit=%d) = %v, want %v", tt.limit, got, tt.want)
			}
			if len(tt.ch) != tt.leftover {
				t.Errorf("после Collect в канале осталось %d значений, want %d (лишнее прочитано?)", len(tt.ch), tt.leftover)
			}
		})
	}
}

func TestCollectNil(t *testing.T) {
	var got []int
	within(t, time.Second, "Collect(nil-канал)", func() { got = Collect(nil, 0) })
	if got != nil {
		t.Errorf("Collect(nil, 0) = %v, want nil", got)
	}
}

// Значения приходят от другой горутины постепенно.
func TestCollectFromGoroutine(t *testing.T) {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for i := range 5 {
			time.Sleep(time.Millisecond)
			ch <- i
		}
	}()
	var got []int
	within(t, time.Second, "Collect", func() { got = Collect(ch, 0) })
	if want := []int{0, 1, 2, 3, 4}; !slices.Equal(got, want) {
		t.Errorf("Collect = %v, want %v", got, want)
	}
}
