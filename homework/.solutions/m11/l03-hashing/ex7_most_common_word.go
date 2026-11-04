package l03hashing

import (
	"strings"
	"unicode"
)

// MostCommonWord возвращает самое частое слово текста, которого нет в banned.
// Слово: максимальная последовательность букв (unicode.IsLetter), всё
// остальное (пробелы, знаки, цифры) разделяет слова. Регистр не важен,
// ответ в нижнем регистре; слова в banned уже в нижнем регистре.
// При равенстве частот верни слово, меньшее при обычном сравнении строк (<).
// Если подходящих слов нет, верни "".
//
// Ожидаемая сложность: O(n) в среднем, где n: длина текста.
func MostCommonWord(text string, banned []string) string {
	ban := make(map[string]bool, len(banned))
	for _, w := range banned {
		ban[w] = true
	}
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r)
	})
	cnt := make(map[string]int)
	for _, w := range words {
		if !ban[w] {
			cnt[w]++
		}
	}
	best, bestCnt := "", 0
	for w, c := range cnt { // порядок обхода map случаен, поэтому явный tie-break
		if c > bestCnt || (c == bestCnt && w < best) {
			best, bestCnt = w, c
		}
	}
	return best
}
