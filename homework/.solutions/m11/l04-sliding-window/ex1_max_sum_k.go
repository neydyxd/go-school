package l04slidingwindow

// MaxSumK возвращает максимальную сумму подотрезка длины ровно k и true.
// Если k <= 0 или k > len(nums), верни 0, false.
// Числа могут быть отрицательными.
//
// Ожидаемая сложность: O(n) по времени. Пересчитывать сумму каждого окна
// заново (O(n*k)) слишком медленно.
func MaxSumK(nums []int, k int) (int, bool) {
	if k <= 0 || k > len(nums) {
		return 0, false
	}
	sum := 0
	for _, v := range nums[:k] {
		sum += v
	}
	best := sum // не 0: все числа могут быть отрицательными
	for r := k; r < len(nums); r++ {
		sum += nums[r] - nums[r-k] // вошёл nums[r], вышел nums[r-k]
		best = max(best, sum)
	}
	return best, true
}
