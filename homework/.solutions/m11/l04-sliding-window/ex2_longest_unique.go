package l04slidingwindow

// LongestUniqueSubstring возвращает длину (в рунах) самой длинной подстроки
// без повторяющихся символов: "abcabcbb" -> 3, "pwwkew" -> 3.
// Пустая строка: 0.
//
// Ожидаемая сложность: O(n) по времени.
func LongestUniqueSubstring(s string) int {
	rs := []rune(s)
	last := make(map[rune]int) // руна -> последняя позиция
	best, l := 0, 0
	for r, c := range rs {
		if p, ok := last[c]; ok && p >= l {
			l = p + 1 // только вперёд: старые позиции левее окна не считаются
		}
		last[c] = r
		best = max(best, r-l+1)
	}
	return best
}
