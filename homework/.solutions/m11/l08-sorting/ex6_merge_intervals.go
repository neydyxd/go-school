package l08sorting

import (
	"cmp"
	"slices"
)

// MergeIntervals объединяет пересекающиеся отрезки [start, end]
// (концы включены, start <= end) и возвращает непересекающиеся отрезки,
// отсортированные по началу. Касающиеся отрезки сливаются: [1 4] и [4 5]
// дают [1 5]. Входной слайс можно переставлять. Пустой вход: пустой ответ.
//
// Ожидаемая сложность: O(n log n). Попарное слияние «пока что-то
// сливается» это O(n^2) и не уложится на большом входе.
func MergeIntervals(intervals [][2]int) [][2]int {
	slices.SortFunc(intervals, func(a, b [2]int) int { return cmp.Compare(a[0], b[0]) })
	var res [][2]int
	for _, iv := range intervals {
		if n := len(res); n > 0 && iv[0] <= res[n-1][1] {
			res[n-1][1] = max(res[n-1][1], iv[1]) // вложенный отрезок не укорачивает текущий
			continue
		}
		res = append(res, iv)
	}
	return res
}
