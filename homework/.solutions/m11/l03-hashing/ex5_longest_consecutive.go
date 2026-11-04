package l03hashing

// LongestConsecutive возвращает длину самой длинной последовательности
// подряд идущих целых чисел (x, x+1, x+2, ...), которые все есть в nums.
// Порядок в nums не важен, повторов может быть много, и они не удлиняют
// последовательность. Значения по модулю не больше 10^9. Пустой вход: 0.
//
// Ожидаемая сложность: O(n) в среднем. Подсказка: начинай отсчёт только
// с тех x, для которых x-1 в наборе нет.
func LongestConsecutive(nums []int) int {
	set := make(map[int]struct{}, len(nums))
	for _, x := range nums {
		set[x] = struct{}{}
	}
	best := 0
	for x := range set {
		if _, ok := set[x-1]; ok {
			continue // x не начало цепочки, её посчитают с начала
		}
		l := 1
		for {
			if _, ok := set[x+l]; !ok {
				break
			}
			l++
		}
		best = max(best, l)
	}
	return best
}
