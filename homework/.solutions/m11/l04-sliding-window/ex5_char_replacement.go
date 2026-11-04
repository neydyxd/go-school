package l04slidingwindow

// CharacterReplacement возвращает длину самой длинной подстроки из одинаковых
// символов (рун), которую можно получить, заменив в s не более k символов
// на любые другие: "AABABBA", k=1 -> 4 ("AABA" -> "AAAA").
// k >= 0. Пустая строка: 0.
//
// Ожидаемая сложность: O(n * A), где A: размер алфавита, или O(n).
func CharacterReplacement(s string, k int) int {
	rs := []rune(s)
	cnt := make(map[rune]int)
	best, maxCnt, l := 0, 0, 0
	for r, c := range rs {
		cnt[c]++
		maxCnt = max(maxCnt, cnt[c])
		// Заменить нужно всё, кроме самого частого символа окна.
		// maxCnt можно не уменьшать при сдвиге l: ответ растёт только
		// вместе с ним, поэтому устаревшее значение ответ не испортит.
		if r-l+1-maxCnt > k {
			cnt[rs[l]]--
			l++
		}
		best = max(best, r-l+1)
	}
	return best
}
