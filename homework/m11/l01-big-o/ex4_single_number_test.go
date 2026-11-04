package l01bigo

import "testing"

func TestSingleNumber(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"один элемент", []int{42}, 42},
		{"ноль одиночка", []int{0, 1, 1}, 0},
		{"классика", []int{4, 1, 2, 1, 2}, 4},
		{"в конце", []int{2, 2, 1}, 1},
		{"отрицательные", []int{-3, 5, -3, -7, 5}, -7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SingleNumber(tt.in); got != tt.want {
				t.Errorf("SingleNumber(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestSingleNumberBig(t *testing.T) {
	const pairs = 200_000
	nums := make([]int, 0, 2*pairs+1)
	for i := range pairs {
		nums = append(nums, i*3-pairs)
	}
	nums = append(nums, 777_777)
	for i := pairs - 1; i >= 0; i-- {
		nums = append(nums, i*3-pairs)
	}
	var got int
	within(t, func() { got = SingleNumber(nums) })
	if got != 777_777 {
		t.Errorf("SingleNumber(большой вход) = %d, want 777777", got)
	}
}
