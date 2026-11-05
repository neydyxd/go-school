package l12dpgreedy

import "sort"

// LengthOfLIS возвращает длину наибольшей строго возрастающей
// подпоследовательности nums. Для пустого слайса ответ 0.
//
// Ожидаемая сложность: O(n log n). Динамика O(n^2) из урока верна,
// но на 200 000 чисел не уложится в лимит: нужен массив «хвостов»
// и бинарный поиск.
func LengthOfLIS(nums []int) int {
	// tails[k] это наименьший возможный последний элемент возрастающей
	// подпоследовательности длины k+1. tails всегда отсортирован.
	var tails []int
	for _, x := range nums {
		// первый хвост >= x: x его заменяет (строгое возрастание)
		i := sort.SearchInts(tails, x)
		if i == len(tails) {
			tails = append(tails, x)
		} else {
			tails[i] = x
		}
	}
	return len(tails)
}
