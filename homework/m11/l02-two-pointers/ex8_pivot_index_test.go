package l02twopointers

import "testing"

func TestPivotIndex(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"пустой", nil, -1},
		{"один элемент", []int{5}, 0},
		{"классика", []int{1, 7, 3, 6, 5, 6}, 3},
		{"нет опоры", []int{1, 2, 3}, -1},
		{"опора в начале", []int{2, 1, -1}, 0},
		{"опора в конце", []int{-1, 1, 7}, 2},
		{"самый левый из нескольких", []int{0, 0, 0}, 0},
		{"отрицательные", []int{-1, -1, -1, -1, -1, 0}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PivotIndex(tt.in); got != tt.want {
				t.Errorf("PivotIndex(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestPivotIndexBig(t *testing.T) {
	const n = 200_000
	nums := make([]int, n)
	for i := range n - 2 {
		nums[i] = 1
	}
	nums[n-2] = 5     // опора: слева n-2 единицы
	nums[n-1] = n - 2 // справа ровно n-2
	var got int
	within(t, func() { got = PivotIndex(nums) })
	if got != n-2 {
		t.Errorf("PivotIndex(большой вход) = %d, want %d", got, n-2)
	}
}
