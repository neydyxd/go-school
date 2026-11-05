package l12dpgreedy

import (
	"cmp"
	"slices"
)

// EraseOverlapIntervals возвращает наименьшее число отрезков, которые
// надо удалить, чтобы оставшиеся попарно не пересекались. Отрезок это
// пара {start, end}, start < end. Отрезки, которые только касаются
// концами ({1, 2} и {2, 3}), не пересекаются. Менять intervals можно.
//
// Ожидаемая сложность: O(n log n), сортировка и один жадный проход.
// Динамика O(n^2) не уложится в лимит на 200 000 отрезков. Подумай,
// по какому концу сортировать: по началу жадность ошибается.
func EraseOverlapIntervals(intervals [][2]int) int {
	// жадно оставляем отрезок, который раньше всех заканчивается:
	// он оставляет больше всего места для остальных
	slices.SortFunc(intervals, func(a, b [2]int) int { return cmp.Compare(a[1], b[1]) })
	kept := 0
	end := 0
	for i, iv := range intervals {
		if i == 0 || iv[0] >= end {
			kept++
			end = iv[1]
		}
	}
	return len(intervals) - kept
}
