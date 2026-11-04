package l02twopointers

import (
	"slices"
	"testing"
)

func TestRemoveDuplicates(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"пустой", []int{}, []int{}},
		{"один", []int{1}, []int{1}},
		{"без повторов", []int{1, 2, 3}, []int{1, 2, 3}},
		{"классика", []int{0, 0, 1, 1, 1, 2, 2, 3, 3, 4}, []int{0, 1, 2, 3, 4}},
		{"все одинаковые", []int{7, 7, 7, 7}, []int{7}},
		{"отрицательные", []int{-3, -3, -1, 0, 0}, []int{-3, -1, 0}},
		{"повтор в конце", []int{1, 2, 2}, []int{1, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nums := slices.Clone(tt.in)
			k := RemoveDuplicates(nums)
			if k < 0 || k > len(nums) || !slices.Equal(nums[:k], tt.want) {
				t.Errorf("RemoveDuplicates(%v): k = %d, nums = %v, want nums[:k] = %v", tt.in, k, nums, tt.want)
			}
		})
	}
}

func TestRemoveDuplicatesBig(t *testing.T) {
	const n = 400_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = i / 4 // каждое значение 4 раза
	}
	var k int
	within(t, func() { k = RemoveDuplicates(nums) })
	if k != n/4 {
		t.Fatalf("RemoveDuplicates(большой вход): k = %d, want %d", k, n/4)
	}
	for i := range k {
		if nums[i] != i {
			t.Fatalf("RemoveDuplicates(большой вход): nums[%d] = %d, want %d", i, nums[i], i)
		}
	}
}
