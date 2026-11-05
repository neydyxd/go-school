package l08sorting

// MergeSortFunc возвращает новый слайс с элементами s, отсортированными
// сортировкой слиянием по функции сравнения cmp (как в slices.SortFunc:
// отрицательное число, если a < b, ноль, если равны, положительное, если
// a > b). Сортировка устойчивая: равные по cmp элементы идут в том же
// порядке, что в s. Исходный слайс s не меняется.
// Пакеты sort и slices для сортировки не используй.
//
// Ожидаемая сложность: O(n log n) по времени, O(n) дополнительной памяти.
func MergeSortFunc[T any](s []T, cmp func(a, b T) int) []T {
	res := make([]T, len(s))
	copy(res, s)
	buf := make([]T, len(s)) // один буфер на все слияния
	mergeSort(res, buf, cmp)
	return res
}

func mergeSort[T any](a, buf []T, cmp func(a, b T) int) {
	if len(a) < 2 {
		return
	}
	mid := len(a) / 2
	mergeSort(a[:mid], buf[:mid], cmp)
	mergeSort(a[mid:], buf[mid:], cmp)
	if cmp(a[mid-1], a[mid]) <= 0 {
		return // половины уже стоят по порядку
	}
	copy(buf, a)
	i, j := 0, mid
	for k := range a {
		// <= 0: при равенстве берём из левой половины, это и есть устойчивость
		if j >= len(a) || (i < mid && cmp(buf[i], buf[j]) <= 0) {
			a[k] = buf[i]
			i++
		} else {
			a[k] = buf[j]
			j++
		}
	}
}
