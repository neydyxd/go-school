package l05stackqueue

// DailyTemperatures по списку дневных температур для каждого дня
// возвращает, сколько дней ждать до первого дня со строго более высокой
// температурой. Если теплее уже не будет, 0.
// [73 74 75 71 69 72 76 73] даёт [1 1 4 2 1 1 0 0].
//
// Ожидаемая сложность: O(n). Перебор вперёд от каждого дня это O(n^2)
// и не уложится, если температура долго не растёт.
func DailyTemperatures(temps []int) []int {
	res := make([]int, len(temps))
	var st []int // индексы дней, которые ещё ждут потепления
	for i, t := range temps {
		for len(st) > 0 && temps[st[len(st)-1]] < t {
			j := st[len(st)-1]
			st = st[:len(st)-1]
			res[j] = i - j
		}
		st = append(st, i)
	}
	return res
}
