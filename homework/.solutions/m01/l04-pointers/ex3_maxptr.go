package l04pointers

// MaxPtr возвращает указатель на наибольший элемент nums (на первый,
// если таких несколько). Указатель должен вести внутрь nums:
// запись *MaxPtr(nums) = 0 меняет сам элемент слайса.
// Для пустого или nil-слайса возвращает nil.
func MaxPtr(nums []int) *int {
	var best *int
	for i := range nums {
		// &nums[i], а не &v из "for _, v := range": v это копия элемента.
		if best == nil || nums[i] > *best {
			best = &nums[i]
		}
	}
	return best
}
