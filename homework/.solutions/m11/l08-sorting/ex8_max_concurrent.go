package l08sorting

import "slices"

// MaxConcurrent по списку сессий [start, end) (начало включено, конец нет,
// start < end, время в секундах до 10^9) возвращает наибольшее число
// сессий, активных одновременно. Сессия [1 5) и [5 8) не пересекаются:
// первая закончилась ровно тогда, когда началась вторая. Пустой вход: 0.
// Входной слайс не меняй.
//
// Ожидаемая сложность: O(n log n): отсортируй отдельно начала и концы
// и пройди их двумя указателями. Попарная проверка O(n^2) и массив на
// все секунды до 10^9 не годятся.
func MaxConcurrent(sessions [][2]int) int {
	starts := make([]int, len(sessions))
	ends := make([]int, len(sessions))
	for i, s := range sessions {
		starts[i], ends[i] = s[0], s[1]
	}
	slices.Sort(starts)
	slices.Sort(ends)
	best, cur, j := 0, 0, 0
	for _, st := range starts {
		for ends[j] <= st { // сессии, закончившиеся к моменту st, уже не активны
			cur--
			j++
		}
		cur++
		best = max(best, cur)
	}
	return best
}
