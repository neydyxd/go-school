package l02twopointers

import "testing"

func TestTwoSumSorted(t *testing.T) {
	tests := []struct {
		name   string
		nums   []int
		target int
		found  bool
	}{
		{"пустой", nil, 5, false},
		{"один элемент", []int{3}, 6, false},
		{"классика", []int{2, 7, 11, 15}, 9, true},
		{"на краях", []int{1, 3, 4, 5, 9}, 10, true},
		{"одинаковые", []int{3, 3}, 6, true},
		{"нельзя брать дважды", []int{1, 3}, 6, false},
		{"отрицательные", []int{-5, -3, 0, 2, 4}, -1, true},
		{"нет пары", []int{1, 2, 3, 4}, 100, false},
		{"повторы", []int{1, 1, 1, 2, 2}, 3, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			i, j := TwoSumSorted(tt.nums, tt.target)
			if !tt.found {
				if i != -1 || j != -1 {
					t.Errorf("TwoSumSorted(%v, %d) = %d, %d, want -1, -1", tt.nums, tt.target, i, j)
				}
				return
			}
			if i < 0 || j >= len(tt.nums) || i >= j || tt.nums[i]+tt.nums[j] != tt.target {
				t.Errorf("TwoSumSorted(%v, %d) = %d, %d: не пара i < j с суммой %d", tt.nums, tt.target, i, j, tt.target)
			}
		})
	}
}

func TestTwoSumSortedBig(t *testing.T) {
	const n = 400_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = 2 * i
	}
	target := nums[n-2] + nums[n-1] // единственная пара: два последних
	var i, j int
	within(t, func() { i, j = TwoSumSorted(nums, target) })
	if i != n-2 || j != n-1 {
		t.Errorf("TwoSumSorted(большой вход, %d) = %d, %d, want %d, %d", target, i, j, n-2, n-1)
	}
}
