package l07maps

import (
	"strings"
	"unicode"
)

// WordCount считает, сколько раз встречается каждое слово в тексте.
//
// Слово: непрерывная последовательность букв и цифр (unicode.IsLetter, unicode.IsDigit),
// всё остальное считается разделителем. Слова приводятся к нижнему регистру.
// Пустой текст: пустая, но не nil мапа (вызывающий код может сразу писать в неё).
func WordCount(text string) map[string]int {
	words := strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	counts := make(map[string]int, len(words))
	for _, w := range words {
		counts[w]++ // нулевое значение int позволяет не проверять наличие ключа
	}
	return counts
}
