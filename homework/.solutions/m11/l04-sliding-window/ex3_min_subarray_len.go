package l04slidingwindow

// MinSubArrayLen возвращает минимальную длину подотрезка nums с суммой
// не меньше target. Все числа nums положительные, target > 0.
// Если такого подотрезка нет, верни 0.
//
// Ожидаемая сложность: O(n) по времени.
func MinSubArrayLen(target int, nums []int) int {
	best := len(nums) + 1
	sum, l := 0, 0
	for r, v := range nums {
		sum += v
		for sum >= target { // окно подходит: пробуем его сузить
			best = min(best, r-l+1)
			sum -= nums[l]
			l++
		}
	}
	if best == len(nums)+1 {
		return 0
	}
	return best
}
