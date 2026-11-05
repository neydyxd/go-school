package l12dpgreedy

// LongestCommonSubsequence возвращает длину наибольшей общей
// подпоследовательности строк a и b, считая в символах (рунах), а не
// в байтах: строки могут быть на русском. Подпоследовательность получается
// вычёркиванием символов без перестановки: "ace" это подпоследовательность
// "abcde". Если общих символов нет, ответ 0.
//
// Ожидаемая сложность: O(n*m) по времени, где n и m длины в рунах.
// Рекурсия без запоминания экспоненциальна.
func LongestCommonSubsequence(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	// dp[i][j] это ответ для префиксов ra[:i] и rb[:j]; хватает двух строк
	prev := make([]int, len(rb)+1)
	cur := make([]int, len(rb)+1)
	for i := 1; i <= len(ra); i++ {
		for j := 1; j <= len(rb); j++ {
			if ra[i-1] == rb[j-1] {
				cur[j] = prev[j-1] + 1
			} else {
				cur[j] = max(prev[j], cur[j-1])
			}
		}
		prev, cur = cur, prev
	}
	return prev[len(rb)]
}
