package l06binarysearch

// FindMinRotated возвращает минимальный элемент сдвинутого отсортированного
// слайса без повторов (см. SearchRotated): [3 4 5 1 2] даёт 1.
// nums не пустой.
//
// Ожидаемая сложность: O(log n). Сравнивай nums[mid] с nums[hi]: это
// говорит, по какую сторону от mid лежит «обрыв».
func FindMinRotated(nums []int) int {
	lo, hi := 0, len(nums)-1
	for lo < hi {
		mid := lo + (hi-lo)/2
		if nums[mid] > nums[hi] {
			lo = mid + 1 // обрыв правее mid, mid точно не минимум
		} else {
			hi = mid // [mid, hi] отсортирован, минимум не правее mid
		}
	}
	return nums[lo]
}
