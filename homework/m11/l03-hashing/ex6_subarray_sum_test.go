package l03hashing

import "testing"

func TestSubarraySum(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		k    int
		want int
	}{
		{"пустой", nil, 0, 0},
		{"классика", []int{1, 1, 1}, 2, 2},
		{"разные", []int{1, 2, 3}, 3, 2},
		{"с отрицательными", []int{1, -1, 0}, 0, 3},
		{"окно не работает", []int{-1, -1, 1}, 0, 1},
		{"один подходит", []int{5}, 5, 1},
		{"нет подходящих", []int{1, 2}, 4, 0},
		{"нули", []int{0, 0, 0}, 0, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SubarraySum(tt.nums, tt.k); got != tt.want {
				t.Errorf("SubarraySum(%v, %d) = %d, want %d", tt.nums, tt.k, got, tt.want)
			}
		})
	}
}

func TestSubarraySumBig(t *testing.T) {
	const n = 400_000
	nums := make([]int, n) // одни нули: подходит каждый из n*(n+1)/2 подотрезков
	var got int
	within(t, func() { got = SubarraySum(nums, 0) })
	if want := n * (n + 1) / 2; got != want {
		t.Errorf("SubarraySum(%d нулей, 0) = %d, want %d", n, got, want)
	}
}
