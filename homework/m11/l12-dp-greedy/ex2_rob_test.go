package l12dpgreedy

import "testing"

func TestRob(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"пустая улица", nil, 0},
		{"один дом", []int{5}, 5},
		{"два дома", []int{2, 7}, 7},
		{"LeetCode 1", []int{1, 2, 3, 1}, 4},
		{"LeetCode 2", []int{2, 7, 9, 3, 1}, 12},
		// чётные-нечётные не работает: выгодно пропустить два дома подряд
		{"пропуск двух подряд", []int{2, 1, 1, 2}, 4},
		{"нули", []int{0, 0, 0}, 0},
		{"жадный по максимуму ошибается", []int{4, 5, 4}, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Rob(tt.in); got != tt.want {
				t.Errorf("Rob(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestRobBig(t *testing.T) {
	const n = 100_000
	nums := make([]int, n)
	for i := range nums {
		nums[i] = 10 + i%3 // 10 11 12 10 11 12 ...
	}
	var got int
	within(t, func() { got = Rob(nums) })
	// эталонный ответ посчитан той же динамикой, что в уроке
	if want := robCheck(nums); got != want {
		t.Errorf("Rob(%d домов) = %d, want %d", n, got, want)
	}
}

// robCheck это простая табличная динамика для проверки ответа.
func robCheck(nums []int) int {
	dp := make([]int, len(nums)+2)
	for i := len(nums) - 1; i >= 0; i-- {
		dp[i] = max(dp[i+1], nums[i]+dp[i+2])
	}
	return dp[0]
}
