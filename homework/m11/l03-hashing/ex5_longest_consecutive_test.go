package l03hashing

import "testing"

func TestLongestConsecutive(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"пустой", nil, 0},
		{"один", []int{7}, 1},
		{"классика", []int{100, 4, 200, 1, 3, 2}, 4},
		{"с повторами", []int{1, 2, 0, 1}, 3},
		{"длинная", []int{0, 3, 7, 2, 5, 8, 4, 6, 0, 1}, 9},
		{"отрицательные", []int{-1, -3, -2, 5}, 3},
		{"все одинаковые", []int{4, 4, 4}, 1},
		{"две цепочки", []int{10, 11, 12, 1, 2}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LongestConsecutive(tt.in); got != tt.want {
				t.Errorf("LongestConsecutive(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestLongestConsecutiveBig(t *testing.T) {
	// Одна цепочка 0..n-1 в перемешанном порядке. Решение, которое
	// начинает подсчёт с каждого числа, сделает около n^2/2 шагов.
	const n = 200_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = (i * 7919) % n
	}
	var got int
	within(t, func() { got = LongestConsecutive(nums) })
	if got != n {
		t.Errorf("LongestConsecutive(перестановка 0..%d) = %d, want %d", n-1, got, n)
	}

	// n/2 повторов начала цепочки и сама цепочка 0..n/2-1. Решение, которое
	// идёт по nums, а не по множеству, пересчитает цепочку для каждого нуля.
	dups := make([]int, 0, n)
	for range n / 2 {
		dups = append(dups, 0)
	}
	for x := 1; x < n/2; x++ {
		dups = append(dups, x)
	}
	within(t, func() { got = LongestConsecutive(dups) })
	if got != n/2 {
		t.Errorf("LongestConsecutive(%d нулей и 1..%d) = %d, want %d", n/2, n/2-1, got, n/2)
	}
}
