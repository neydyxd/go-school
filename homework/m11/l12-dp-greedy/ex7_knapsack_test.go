package l12dpgreedy

import "testing"

func TestKnapsack(t *testing.T) {
	tests := []struct {
		name     string
		w, v     []int
		capacity int
		want     int
	}{
		{"нет предметов", nil, nil, 10, 0},
		{"нулевая вместимость", []int{1}, []int{5}, 0, 0},
		{"классика", []int{1, 3, 4, 5}, []int{1, 4, 5, 7}, 7, 9},
		{"предмет не влезает", []int{5}, []int{10}, 4, 0},
		// если брать предмет многократно, вышло бы 3
		{"каждый не больше раза", []int{1}, []int{1}, 3, 1},
		// жадно по удельной ценности: 6/5 > 5/5, но 5+5 лучше
		{"жадный по удельной ценности", []int{6, 5, 5}, []int{7, 5, 5}, 10, 10},
		{"бесполезный предмет", []int{2, 3}, []int{0, 4}, 5, 4},
		{"ровно вместимость", []int{2, 3, 4}, []int{3, 4, 5}, 5, 7},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Knapsack(tt.w, tt.v, tt.capacity); got != tt.want {
				t.Errorf("Knapsack(%v, %v, %d) = %d, want %d", tt.w, tt.v, tt.capacity, got, tt.want)
			}
		})
	}
}

func TestKnapsackBig(t *testing.T) {
	// 200 предметов: вес i+1, ценность (i*37)%101 + 1, вместимость 10 000
	const n, capacity = 200, 10_000
	w, v := make([]int, n), make([]int, n)
	for i := range n {
		w[i], v[i] = i+1, (i*37)%101+1
	}
	var got int
	within(t, func() { got = Knapsack(w, v, capacity) })
	if want := knapsackCheck(w, v, capacity); got != want {
		t.Errorf("Knapsack(200 предметов, %d) = %d, want %d", capacity, got, want)
	}
}

// knapsackCheck это двумерная таблица dp[i][c] для проверки ответа.
func knapsackCheck(w, v []int, capacity int) int {
	dp := make([][]int, len(w)+1)
	for i := range dp {
		dp[i] = make([]int, capacity+1)
	}
	for i := 1; i <= len(w); i++ {
		for c := 0; c <= capacity; c++ {
			dp[i][c] = dp[i-1][c]
			if c >= w[i-1] {
				dp[i][c] = max(dp[i][c], dp[i-1][c-w[i-1]]+v[i-1])
			}
		}
	}
	return dp[len(w)][capacity]
}
