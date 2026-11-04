package l03functions

// Sum возвращает сумму всех аргументов. Без аргументов: 0.
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// MinMax возвращает наименьший и наибольший из аргументов.
// Первый аргумент обязателен (сигнатура не даёт вызвать MinMax() без аргументов),
// остальных может не быть: MinMax(5) == (5, 5).
func MinMax(first int, rest ...int) (lo, hi int) {
	lo, hi = first, first
	for _, n := range rest {
		lo = min(lo, n)
		hi = max(hi, n)
	}
	return lo, hi
}
