package l12dpgreedy

import "testing"

func TestChangeWays(t *testing.T) {
	tests := []struct {
		name   string
		amount int
		coins  []int
		want   int
	}{
		{"LeetCode 1", 5, []int{1, 2, 5}, 4},
		{"нельзя набрать", 3, []int{2}, 0},
		{"одна монета", 10, []int{10}, 1},
		{"сумма ноль", 0, []int{7}, 1},
		{"сумма ноль без монет", 0, nil, 1},
		{"без монет", 5, nil, 0},
		// перестановок 1+2, 2+1, 1+1+1 было бы 3, наборов 2
		{"порядок не важен", 3, []int{1, 2}, 2},
		{"монета больше суммы", 4, []int{1, 10}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChangeWays(tt.amount, tt.coins); got != tt.want {
				t.Errorf("ChangeWays(%d, %v) = %d, want %d", tt.amount, tt.coins, got, tt.want)
			}
		})
	}
}

func TestChangeWaysBig(t *testing.T) {
	var got int
	within(t, func() { got = ChangeWays(10_000, []int{1, 2, 5}) })
	if want := 5_004_001; got != want {
		t.Errorf("ChangeWays(10000, [1 2 5]) = %d, want %d", got, want)
	}
}
