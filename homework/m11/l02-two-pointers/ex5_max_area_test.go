package l02twopointers

import "testing"

func TestMaxArea(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"пустой", nil, 0},
		{"один отрезок", []int{5}, 0},
		{"два", []int{1, 1}, 1},
		{"классика", []int{1, 8, 6, 2, 5, 4, 8, 3, 7}, 49},
		{"высокие в середине", []int{1, 2, 100, 100, 2, 1}, 100},
		{"нули", []int{0, 0, 0}, 0},
		{"убывающие", []int{5, 4, 3, 2, 1}, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxArea(tt.in); got != tt.want {
				t.Errorf("MaxArea(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestMaxAreaBig(t *testing.T) {
	const n = 200_000
	heights := make([]int, n)
	for i := range heights {
		heights[i] = 1
	}
	heights[10] = 1_000_000
	heights[n-11] = 1_000_000
	var got int
	within(t, func() { got = MaxArea(heights) })
	if want := (n - 21) * 1_000_000; got != want {
		t.Errorf("MaxArea(большой вход) = %d, want %d", got, want)
	}
}
