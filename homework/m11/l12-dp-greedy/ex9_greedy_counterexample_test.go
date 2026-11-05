package l12dpgreedy

import "testing"

func TestGreedyCounterexample(t *testing.T) {
	tests := []struct {
		name  string
		coins []int
		want  int
	}{
		{"одна монета", []int{1}, -1},
		{"рубли", []int{1, 2, 5, 10, 50, 100, 200, 500, 1000, 2000, 5000}, -1},
		{"центы США", []int{1, 5, 10, 25}, -1},
		{"классический контрпример", []int{1, 3, 4}, 6},
		{"монета в 20", []int{1, 5, 10, 20, 25}, 40},
		{"две монеты всегда жадные", []int{1, 7}, -1},
		{"контрпример у границы", []int{1, 10, 11}, 20},
		{"британские до 1971", []int{1, 3, 6, 12, 24, 30}, 48},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// даже на маленьких наборах перебор без запоминания идёт часами,
			// поэтому и здесь ограничиваем время
			var got int
			within(t, func() { got = GreedyCounterexample(tt.coins) })
			if got != tt.want {
				t.Errorf("GreedyCounterexample(%v) = %d, want %d", tt.coins, got, tt.want)
			}
		})
	}
}

func TestGreedyCounterexampleBig(t *testing.T) {
	// граница почти 2*10^5, монет 300: таблица на 6*10^7 шагов
	coins := []int{1}
	for c := 333; c < 100_000; c += 333 {
		coins = append(coins, c)
	}
	coins = append(coins, 99_999, 100_000)
	var got int
	within(t, func() { got = GreedyCounterexample(coins) })
	if want := greedyCheck(coins); got != want {
		t.Errorf("GreedyCounterexample(%d монет до 100000) = %d, want %d", len(coins), got, want)
	}
}

// greedyCheck: та же проверка простым перебором для сверки ответа.
func greedyCheck(coins []int) int {
	n := len(coins)
	if n < 2 {
		return -1
	}
	limit := coins[n-2] + coins[n-1]
	best := make([]int, limit)
	for s := 1; s < limit; s++ {
		best[s] = s
		for _, c := range coins {
			if c <= s {
				best[s] = min(best[s], best[s-c]+1)
			}
		}
		g, rest := 0, s
		for i := n - 1; i >= 0; i-- {
			g += rest / coins[i]
			rest %= coins[i]
		}
		if g > best[s] {
			return s
		}
	}
	return -1
}
