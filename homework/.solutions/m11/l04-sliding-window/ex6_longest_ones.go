package l04slidingwindow

// LongestOnes получает слайс из нулей и единиц и число k >= 0. Можно
// превратить не более k нулей в единицы. Верни длину самого длинного
// подотрезка из одних единиц, который после этого получится.
// Пустой слайс: 0.
//
// Ожидаемая сложность: O(n) по времени.
func LongestOnes(nums []int, k int) int {
	best, zeros, l := 0, 0, 0
	for r, v := range nums {
		if v == 0 {
			zeros++
		}
		for zeros > k {
			if nums[l] == 0 {
				zeros--
			}
			l++
		}
		best = max(best, r-l+1)
	}
	return best
}
