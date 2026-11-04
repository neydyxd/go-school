package l02twopointers

import (
	"slices"
	"testing"
)

func TestMoveZeroes(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"пустой", []int{}, []int{}},
		{"только ноль", []int{0}, []int{0}},
		{"классика", []int{0, 1, 0, 3, 12}, []int{1, 3, 12, 0, 0}},
		{"без нулей", []int{4, 2, 7}, []int{4, 2, 7}},
		{"все нули", []int{0, 0, 0}, []int{0, 0, 0}},
		{"порядок и отрицательные", []int{-1, 0, 5, 0, -2, 0}, []int{-1, 5, -2, 0, 0, 0}},
		{"нули уже в конце", []int{1, 2, 0, 0}, []int{1, 2, 0, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nums := slices.Clone(tt.in)
			MoveZeroes(nums)
			if !slices.Equal(nums, tt.want) {
				t.Errorf("MoveZeroes(%v) = %v, want %v", tt.in, nums, tt.want)
			}
		})
	}
}

func TestMoveZeroesBig(t *testing.T) {
	const n = 400_000
	nums := make([]int, n)
	for i := range nums {
		if i%2 == 1 {
			nums[i] = i
		}
	}
	within(t, func() { MoveZeroes(nums) })
	for i := range n / 2 {
		if nums[i] != 2*i+1 {
			t.Fatalf("MoveZeroes(большой вход): nums[%d] = %d, want %d", i, nums[i], 2*i+1)
		}
	}
	for i := n / 2; i < n; i++ {
		if nums[i] != 0 {
			t.Fatalf("MoveZeroes(большой вход): nums[%d] = %d, want 0", i, nums[i])
		}
	}
}
