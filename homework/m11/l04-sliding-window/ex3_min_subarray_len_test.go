package l04slidingwindow

import "testing"

func TestMinSubArrayLen(t *testing.T) {
	tests := []struct {
		name   string
		target int
		nums   []int
		want   int
	}{
		{"пустой", 3, nil, 0},
		{"классика", 7, []int{2, 3, 1, 2, 4, 3}, 2},
		{"один элемент хватает", 4, []int{1, 4, 4}, 1},
		{"не набрать", 11, []int{1, 1, 1, 1, 1, 1, 1, 1}, 0},
		{"весь массив", 15, []int{1, 2, 3, 4, 5}, 5},
		{"ровно target", 6, []int{1, 2, 3}, 3},
		{"окно в конце", 10, []int{1, 1, 1, 1, 5, 5}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MinSubArrayLen(tt.target, tt.nums); got != tt.want {
				t.Errorf("MinSubArrayLen(%d, %v) = %d, want %d", tt.target, tt.nums, got, tt.want)
			}
		})
	}
}

func TestMinSubArrayLenBig(t *testing.T) {
	const n = 400_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = 1
	}
	var got int
	within(t, func() { got = MinSubArrayLen(n/2, nums) })
	if got != n/2 {
		t.Errorf("MinSubArrayLen(%d, %d единиц) = %d, want %d", n/2, n, got, n/2)
	}
}
