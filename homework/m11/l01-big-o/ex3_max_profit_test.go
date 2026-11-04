package l01bigo

import "testing"

func TestMaxProfit(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"пустой", nil, 0},
		{"один день", []int{5}, 0},
		{"классика", []int{7, 1, 5, 3, 6, 4}, 5},
		{"только падение", []int{7, 6, 4, 3, 1}, 0},
		{"минимум после максимума", []int{3, 8, 1}, 5},
		{"минимум в конце не помогает", []int{2, 4, 1}, 2},
		{"новый минимум и рост после", []int{3, 9, 1, 8}, 7},
		{"одинаковые цены", []int{4, 4, 4}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxProfit(tt.in); got != tt.want {
				t.Errorf("MaxProfit(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestMaxProfitBig(t *testing.T) {
	const n = 200_000
	prices := make([]int, n)
	for i := range prices {
		prices[i] = n - i // всё время падает, прибыли нет
	}
	prices[n/2] = 1
	prices[n-1] = 1_000_000
	var got int
	within(t, func() { got = MaxProfit(prices) })
	if want := 1_000_000 - 1; got != want {
		t.Errorf("MaxProfit(большой вход) = %d, want %d", got, want)
	}
}
