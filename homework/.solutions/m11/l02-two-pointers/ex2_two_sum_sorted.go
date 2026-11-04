package l02twopointers

// TwoSumSorted получает слайс, отсортированный по неубыванию, и target.
// Верни индексы i < j, такие что nums[i]+nums[j] == target. Если подходящих
// пар несколько, верни любую. Если пары нет, верни -1, -1.
// Один и тот же элемент дважды брать нельзя.
//
// Ожидаемая сложность: O(n) по времени и O(1) по дополнительной памяти.
func TwoSumSorted(nums []int, target int) (int, int) {
	l, r := 0, len(nums)-1
	for l < r {
		switch sum := nums[l] + nums[r]; {
		case sum == target:
			return l, r
		case sum < target:
			l++ // нужна сумма больше: двигаем левый к большим числам
		default:
			r-- // нужна сумма меньше
		}
	}
	return -1, -1
}
