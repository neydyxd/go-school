package l04slidingwindow

import "testing"

func TestLongestOnes(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		k    int
		want int
	}{
		{"пустой", nil, 1, 0},
		{"классика", []int{1, 1, 1, 0, 0, 0, 1, 1, 1, 1, 0}, 2, 6},
		{"длинная", []int{0, 0, 1, 1, 0, 0, 1, 1, 1, 0, 1, 1, 0, 0, 0, 1, 1, 1, 1}, 3, 10},
		{"k=0", []int{1, 0, 1, 1, 0, 1}, 0, 2},
		{"одни нули без замен", []int{0, 0, 0}, 0, 0},
		{"одни нули с заменами", []int{0, 0, 0}, 2, 2},
		{"k больше числа нулей", []int{0, 1, 0}, 5, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LongestOnes(tt.nums, tt.k); got != tt.want {
				t.Errorf("LongestOnes(%v, %d) = %d, want %d", tt.nums, tt.k, got, tt.want)
			}
		})
	}
}

func TestLongestOnesBig(t *testing.T) {
	const n = 200_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = (i + 1) % 2 // 1 0 1 0 ...
	}
	const k = 50_000
	var got int
	within(t, func() { got = LongestOnes(nums, k) })
	if want := 2*k + 1; got != want {
		t.Errorf("LongestOnes(1 0 1 0 ..., %d) = %d, want %d", k, got, want)
	}
}
