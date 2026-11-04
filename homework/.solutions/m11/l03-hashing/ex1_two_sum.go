package l03hashing

// TwoSum получает НЕотсортированный слайс и target. Верни индексы i < j,
// такие что nums[i]+nums[j] == target. Если подходящих пар несколько,
// верни любую. Если пары нет, верни -1, -1. Один и тот же элемент дважды
// брать нельзя.
//
// Ожидаемая сложность: O(n) по времени и O(n) по памяти.
func TwoSum(nums []int, target int) (int, int) {
	seen := make(map[int]int, len(nums)) // значение -> индекс
	for j, x := range nums {
		if i, ok := seen[target-x]; ok {
			return i, j
		}
		seen[x] = j // кладём после проверки: так элемент не сложится сам с собой
	}
	return -1, -1
}
