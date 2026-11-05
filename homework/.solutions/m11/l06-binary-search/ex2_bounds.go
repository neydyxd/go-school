package l06binarysearch

// LowerBound возвращает первый индекс i, для которого nums[i] >= x,
// а UpperBound первый индекс i, для которого nums[i] > x. Если такого
// индекса нет, обе возвращают len(nums). nums отсортирован по
// неубыванию, повторы возможны. [1 2 2 2 5], x = 2: LowerBound = 1,
// UpperBound = 4, а UpperBound - LowerBound это число двоек.
// Напиши руками, без пакетов sort и slices.
//
// Ожидаемая сложность: O(log n) каждая.
func LowerBound(nums []int, x int) int {
	return firstTrue(len(nums), func(i int) bool { return nums[i] >= x })
}

func UpperBound(nums []int, x int) int {
	return firstTrue(len(nums), func(i int) bool { return nums[i] > x })
}

// firstTrue возвращает первый i из [0, n), где pred(i) истинно, или n.
// pred монотонен: false...false true...true.
func firstTrue(n int, pred func(int) bool) int {
	lo, hi := 0, n // ответ лежит в [lo, hi]
	for lo < hi {
		mid := lo + (hi-lo)/2
		if pred(mid) {
			hi = mid // mid подходит, но может есть левее
		} else {
			lo = mid + 1
		}
	}
	return lo
}
