package l04slidingwindow

// NumSubarrayProductLessThanK возвращает, сколько непустых подотрезков
// nums имеют произведение элементов строго меньше k.
// Все числа nums положительные. Если k <= 1, ответ 0.
//
// Ожидаемая сложность: O(n) по времени. Ответ может быть до n*(n+1)/2.
// Числа и k могут быть огромными (до ~9*10^18), и произведение двух
// элементов легко переполнит int: прежде чем умножать окно на nums[r],
// проверь делением, что результат останется меньше k.
func NumSubarrayProductLessThanK(nums []int, k int) int {
	panic("TODO: реализуй NumSubarrayProductLessThanK")
}
