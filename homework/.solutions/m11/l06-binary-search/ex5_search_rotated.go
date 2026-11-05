package l06binarysearch

// SearchRotated ищет target в слайсе, который был отсортирован по
// возрастанию без повторов, а потом циклически сдвинут: [4 5 6 7 0 1 2].
// Сдвиг может быть нулевым. Возвращает индекс target или -1.
// Напиши бинарный поиск руками, без пакетов sort и slices.
//
// Ожидаемая сложность: O(log n). Подсказка: при любом mid хотя бы одна
// из половин [lo, mid] и [mid, hi] отсортирована, и по её концам видно,
// лежит ли target внутри.
func SearchRotated(nums []int, target int) int {
	lo, hi := 0, len(nums)-1
	for lo <= hi {
		mid := lo + (hi-lo)/2
		if nums[mid] == target {
			return mid
		}
		if nums[lo] <= nums[mid] { // левая половина отсортирована
			if nums[lo] <= target && target < nums[mid] {
				hi = mid - 1
			} else {
				lo = mid + 1
			}
		} else { // отсортирована правая половина
			if nums[mid] < target && target <= nums[hi] {
				lo = mid + 1
			} else {
				hi = mid - 1
			}
		}
	}
	return -1
}
