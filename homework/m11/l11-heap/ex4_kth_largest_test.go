package l11heap

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestFindKthLargest(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		k    int
		want int
	}{
		{"один элемент", []int{1}, 1, 1},
		{"LeetCode 1", []int{3, 2, 1, 5, 6, 4}, 2, 5},
		{"LeetCode 2", []int{3, 2, 3, 1, 2, 4, 5, 5, 6}, 4, 4},
		{"повторы считаются", []int{3, 3, 1}, 2, 3},
		{"k = n это минимум", []int{4, -1, 7, 0}, 4, -1},
		{"k = 1 это максимум", []int{4, -1, 7, 0}, 1, 7},
		{"отрицательные", []int{-5, -1, -3}, 2, -3},
		{"края int", []int{math.MinInt, math.MaxInt, 0}, 3, math.MinInt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindKthLargest(slices.Clone(tt.nums), tt.k); got != tt.want {
				t.Errorf("FindKthLargest(%v, %d) = %d, want %d", tt.nums, tt.k, got, tt.want)
			}
		})
	}
}

func TestFindKthLargestBig(t *testing.T) {
	const n = 500_000
	r := rand.New(rand.NewPCG(19, 20))
	nums := make([]int, n)
	for i := range nums {
		nums[i] = r.IntN(1 << 30)
	}
	sorted := slices.Sorted(slices.Values(nums))
	for _, k := range []int{1, 50_000, n / 2} {
		want := sorted[n-k]
		var got int
		in := slices.Clone(nums)
		within(t, func() { got = FindKthLargest(in, k) })
		if got != want {
			t.Errorf("FindKthLargest(%d случайных, %d) = %d, want %d", n, k, got, want)
		}
	}
}
