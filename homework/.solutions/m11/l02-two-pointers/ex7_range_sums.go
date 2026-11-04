package l02twopointers

// RangeSums отвечает на запросы сумм на отрезках: для каждого запроса
// [l, r] верни nums[l] + nums[l+1] + ... + nums[r] (обе границы включены).
// Гарантируется 0 <= l <= r < len(nums). Ответы в порядке запросов.
// Нет запросов: пустой результат (или nil).
//
// Ожидаемая сложность: O(n + q), где q = len(queries). Считать каждую сумму
// циклом заново (O(n*q)) слишком медленно.
func RangeSums(nums []int, queries [][2]int) []int {
	// prefix[i] = сумма первых i элементов, prefix[0] = 0.
	prefix := make([]int, len(nums)+1)
	for i, v := range nums {
		prefix[i+1] = prefix[i] + v
	}
	res := make([]int, len(queries))
	for i, q := range queries {
		res[i] = prefix[q[1]+1] - prefix[q[0]]
	}
	return res
}
