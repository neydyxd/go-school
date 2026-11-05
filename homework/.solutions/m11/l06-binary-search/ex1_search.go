package l06binarysearch

// Search ищет target в отсортированном по возрастанию слайсе nums
// без повторов и возвращает его индекс или -1, если его нет.
// Напиши бинарный поиск руками, без пакетов sort и slices.
//
// Ожидаемая сложность: O(log n). Линейный проход не уложится в тесте
// с сотнями тысяч запросов.
func Search(nums []int, target int) int {
	lo, hi := 0, len(nums)-1 // ищем на отрезке [lo, hi] включительно
	for lo <= hi {
		mid := lo + (hi-lo)/2
		switch {
		case nums[mid] == target:
			return mid
		case nums[mid] < target:
			lo = mid + 1
		default:
			hi = mid - 1
		}
	}
	return -1
}
