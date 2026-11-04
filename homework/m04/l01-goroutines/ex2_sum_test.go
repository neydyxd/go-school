package l01goroutines

import (
	"testing"
	"time"
)

func TestConcurrentSum(t *testing.T) {
	big := make([]int, 10_000)
	bigSum := 0
	for i := range big {
		big[i] = i - 3000
		bigSum += big[i]
	}
	tests := []struct {
		name  string
		nums  []int
		parts int
		want  int
	}{
		{"nil", nil, 4, 0},
		{"пустой", []int{}, 4, 0},
		{"parts = 0 как один кусок", []int{1, 2, 3}, 0, 6},
		{"отрицательный parts", []int{1, 2, 3}, -5, 6},
		{"один кусок", []int{1, 2, 3, 4}, 1, 10},
		{"куски неравные", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 3, 55},
		{"parts больше длины", []int{5, 5}, 100, 10},
		{"отрицательные числа", []int{-1, -2, 3}, 2, 0},
		{"большой слайс", big, 8, bigSum},
		{"большой слайс, кусков больше ядер", big, 333, bigSum},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			within(t, 2*time.Second, "ConcurrentSum", func() { got = ConcurrentSum(tt.nums, tt.parts) })
			if got != tt.want {
				t.Errorf("ConcurrentSum(len=%d, parts=%d) = %d, want %d", len(tt.nums), tt.parts, got, tt.want)
			}
		})
	}
}
