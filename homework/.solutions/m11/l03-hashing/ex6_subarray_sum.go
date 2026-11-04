package l03hashing

// SubarraySum возвращает, сколько непустых подотрезков nums (непрерывных
// кусков) имеют сумму ровно k. Числа могут быть отрицательными и нулевыми,
// поэтому скользящее окно здесь не работает.
//
// Ожидаемая сложность: O(n) в среднем (префиксные суммы + map).
// Ответ может быть большим: до n*(n+1)/2.
func SubarraySum(nums []int, k int) int {
	// seen[p]: сколько префиксов (включая пустой) имеют сумму p.
	// Подотрезок (i, j] имеет сумму k, если prefix[j] - prefix[i] == k.
	seen := map[int]int{0: 1}
	count, prefix := 0, 0
	for _, x := range nums {
		prefix += x
		count += seen[prefix-k]
		seen[prefix]++
	}
	return count
}
