package l01bigo

import (
	"slices"
	"testing"
)

func TestRotate(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		k    int
		want []int
	}{
		{"пустой", []int{}, 3, []int{}},
		{"k=0", []int{1, 2, 3}, 0, []int{1, 2, 3}},
		{"классика", []int{1, 2, 3, 4, 5, 6, 7}, 3, []int{5, 6, 7, 1, 2, 3, 4}},
		{"k равно длине", []int{1, 2, 3}, 3, []int{1, 2, 3}},
		{"k больше длины", []int{1, 2, 3}, 4, []int{3, 1, 2}},
		{"один элемент", []int{9}, 100, []int{9}},
		{"отрицательные значения", []int{-1, -100, 3, 99}, 2, []int{3, 99, -1, -100}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nums := slices.Clone(tt.in)
			Rotate(nums, tt.k)
			if !slices.Equal(nums, tt.want) {
				t.Errorf("Rotate(%v, %d) = %v, want %v", tt.in, tt.k, nums, tt.want)
			}
		})
	}
}

func TestRotateBig(t *testing.T) {
	const n = 200_000
	const k = 100_007
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i
	}
	within(t, func() { Rotate(nums, k) })
	for i, v := range nums {
		if want := (i - k + n) % n; v != want {
			t.Fatalf("Rotate(0..%d, %d): nums[%d] = %d, want %d", n-1, k, i, v, want)
		}
	}
}
