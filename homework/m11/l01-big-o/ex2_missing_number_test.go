package l01bigo

import "testing"

func TestMissingNumber(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"пустой", []int{}, 0},
		{"нет нуля", []int{1}, 0},
		{"нет единицы", []int{0}, 1},
		{"нет последнего", []int{0, 1}, 2},
		{"в середине", []int{3, 0, 1}, 2},
		{"длинный", []int{9, 6, 4, 2, 3, 5, 7, 0, 1}, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MissingNumber(tt.in); got != tt.want {
				t.Errorf("MissingNumber(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestMissingNumberBig(t *testing.T) {
	const n = 200_000
	const missing = 123_456
	nums := make([]int, 0, n)
	for i := n; i >= 0; i-- {
		if i != missing {
			nums = append(nums, i)
		}
	}
	var got int
	within(t, func() { got = MissingNumber(nums) })
	if got != missing {
		t.Errorf("MissingNumber(0..%d без %d) = %d, want %d", n, missing, got, missing)
	}
}
