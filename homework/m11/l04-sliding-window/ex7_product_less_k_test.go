package l04slidingwindow

import "testing"

func TestNumSubarrayProductLessThanK(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		k    int
		want int
	}{
		{"пустой", nil, 10, 0},
		{"классика", []int{10, 5, 2, 6}, 100, 8},
		{"k ноль", []int{1, 2, 3}, 0, 0},
		{"k один", []int{1, 1, 1}, 1, 0},
		{"все подходят", []int{1, 2, 3}, 100, 6},
		{"ни один не подходит", []int{5, 6}, 5, 0},
		{"большие числа", []int{1_000_000_000, 1_000_000_000, 1_000_000_000, 1}, 1_000_000_001, 5},
		{"произведение пары не влезает в int", []int{3_000_000_000, 4_000_000_000}, 10_000_000_000, 2},
		{"огромные k и элемент", []int{1_000_000_000_000_000_000, 10}, 2_000_000_000_000_000_000, 2},
		{"элемент не меньше k рвёт окно", []int{2, 100, 3}, 50, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NumSubarrayProductLessThanK(tt.nums, tt.k); got != tt.want {
				t.Errorf("NumSubarrayProductLessThanK(%v, %d) = %d, want %d", tt.nums, tt.k, got, tt.want)
			}
		})
	}
}

func TestNumSubarrayProductLessThanKBig(t *testing.T) {
	const n = 200_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = 1
	}
	var got int
	within(t, func() { got = NumSubarrayProductLessThanK(nums, 2) })
	if want := n * (n + 1) / 2; got != want {
		t.Errorf("NumSubarrayProductLessThanK(%d единиц, 2) = %d, want %d", n, got, want)
	}
}
