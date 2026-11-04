package l01bigo

// MaxProfit получает цены акции по дням. Можно один раз купить и позже
// (в другой день) один раз продать. Верни максимальную прибыль; если
// заработать нельзя, верни 0. Пустой слайс и один день: 0.
//
// Ожидаемая сложность: O(n) по времени, O(1) по памяти.
func MaxProfit(prices []int) int {
	best := 0
	minSoFar := 0 // минимальная цена среди уже просмотренных дней
	for i, p := range prices {
		if i == 0 || p < minSoFar {
			minSoFar = p
			continue
		}
		best = max(best, p-minSoFar)
	}
	return best
}
