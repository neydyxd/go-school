package l02channels

import (
	"fmt"
	"runtime"
	"testing"
	"time"
)

func TestBatch(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		size int
		want [][]int
	}{
		{"пустой вход", nil, 3, nil},
		{"ровно пачками", []int{1, 2, 3, 4}, 2, [][]int{{1, 2}, {3, 4}}},
		{"хвост", []int{1, 2, 3, 4, 5}, 2, [][]int{{1, 2}, {3, 4}, {5}}},
		{"одна неполная", []int{1, 2}, 5, [][]int{{1, 2}}},
		{"size = 1", []int{7, 8}, 1, [][]int{{7}, {8}}},
		{"size = 0 как 1", []int{7, 8}, 0, [][]int{{7}, {8}}},
		{"отрицательный size как 1", []int{9}, -3, [][]int{{9}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := runtime.NumGoroutine()
			in := make(chan int)
			go func() {
				defer close(in)
				for _, v := range tt.in {
					in <- v
				}
			}()
			var out <-chan []int
			within(t, time.Second, "Batch (должна вернуть канал сразу)", func() { out = Batch(in, tt.size) })
			// Сначала собираем все пачки и только потом сравниваем:
			// так видно, если новая пачка перетёрла старую.
			var got [][]int
			within(t, time.Second, "чтение пачек (выходной канал закрыт?)", func() {
				for b := range out {
					got = append(got, b)
				}
			})
			if fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Errorf("Batch(%v, %d) = %v, want %v", tt.in, tt.size, got, tt.want)
			}
			noLeak(t, before)
		})
	}
}
