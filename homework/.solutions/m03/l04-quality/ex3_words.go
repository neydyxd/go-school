package l04quality

import (
	"strings"
	"unicode"
)

// CountWords считает слова в тексте без учёта регистра.
func CountWords(text string) map[string]int {
	counts := map[string]int{}
	for _, w := range strings.FieldsFunc(text, notWordRune) {
		counts[strings.ToLower(w)]++
	}
	return counts
}

func notWordRune(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsNumber(r)
}
