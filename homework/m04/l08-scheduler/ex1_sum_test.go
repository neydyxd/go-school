package l08scheduler

import (
	"math"
	"testing"
)

func seq(n int) []int64 {
	s := make([]int64, n)
	for i := range s {
		s[i] = int64(i + 1)
	}
	return s
}

func TestParallelSum(t *testing.T) {
	tests := []struct {
		name  string
		nums  []int64
		parts int
		want  int64
	}{
		{"пусто", nil, 4, 0},
		{"один элемент", []int64{42}, 8, 42},
		{"делится нацело", seq(100), 4, 5050},
		{"не делится нацело", seq(10), 3, 55},
		{"частей больше, чем элементов", seq(3), 10, 6},
		{"parts=0: GOMAXPROCS", seq(1000), 0, 500500},
		{"отрицательный parts", seq(7), -1, 28},
		{"одна часть", seq(5), 1, 15},
		{"отрицательные числа", []int64{-5, 10, -3, 8}, 2, 10},
		{"большие числа", []int64{math.MaxInt32, math.MaxInt32, math.MaxInt32}, 3, 3 * math.MaxInt32},
		{"миллион", seq(1_000_000), 7, 500000500000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParallelSum(tt.nums, tt.parts); got != tt.want {
				t.Errorf("ParallelSum(%d чисел, %d) = %d, want %d", len(tt.nums), tt.parts, got, tt.want)
			}
		})
	}
}
