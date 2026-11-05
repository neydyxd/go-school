package l12dpgreedy

import "testing"

func TestCoinChange(t *testing.T) {
	tests := []struct {
		name   string
		coins  []int
		amount int
		want   int
	}{
		{"LeetCode 1", []int{1, 2, 5}, 11, 3},
		{"нельзя набрать", []int{2}, 3, -1},
		{"сумма ноль", []int{1}, 0, 0},
		{"контрпример жадному", []int{1, 3, 4}, 6, 2},
		{"жадный застревает", []int{5, 3}, 9, 3},
		{"одна монета", []int{7}, 14, 2},
		{"порядок монет не важен", []int{5, 1, 2}, 11, 3},
		{"монета больше суммы", []int{10, 3}, 6, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CoinChange(tt.coins, tt.amount); got != tt.want {
				t.Errorf("CoinChange(%v, %d) = %d, want %d", tt.coins, tt.amount, got, tt.want)
			}
		})
	}
}

func TestCoinChangeBig(t *testing.T) {
	tests := []struct {
		coins  []int
		amount int
		want   int
	}{
		{[]int{186, 419, 83, 408}, 6249, 20},
		{[]int{2, 5, 10, 1}, 27, 4},
		{[]int{7, 11}, 100_000, 9092},
		{[]int{4, 6}, 100_001, -1}, // нечётную сумму чётными не набрать
	}
	for _, tt := range tests {
		var got int
		within(t, func() { got = CoinChange(tt.coins, tt.amount) })
		if got != tt.want {
			t.Errorf("CoinChange(%v, %d) = %d, want %d", tt.coins, tt.amount, got, tt.want)
		}
	}
}
