package l07maps

import (
	"cmp"
	"maps"
	"slices"
)

// TopN возвращает n самых частых слов из freq.
//
// Порядок: по убыванию частоты, при равной частоте по алфавиту (обычное сравнение строк).
// Результат обязан быть одинаковым при каждом вызове, хотя порядок обхода мапы случаен.
// n больше числа слов: вернуть все. n <= 0 или пустая мапа: слайс длины 0.
func TopN(freq map[string]int, n int) []string {
	if n <= 0 {
		return []string{}
	}
	words := slices.Collect(maps.Keys(freq))
	slices.SortFunc(words, func(a, b string) int {
		if c := cmp.Compare(freq[b], freq[a]); c != 0 {
			return c // по убыванию частоты
		}
		return cmp.Compare(a, b)
	})
	if n > len(words) {
		n = len(words)
	}
	return words[:n]
}
