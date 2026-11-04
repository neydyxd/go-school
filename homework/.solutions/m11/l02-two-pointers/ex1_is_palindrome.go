package l02twopointers

import (
	"unicode"
	"unicode/utf8"
)

// IsPalindrome сообщает, читается ли строка одинаково в обе стороны,
// если учитывать только буквы и цифры (unicode.IsLetter, unicode.IsDigit)
// и не различать регистр (unicode.ToLower).
// Строка без букв и цифр, в том числе пустая, считается палиндромом.
//
// Строка может быть на любом языке: сравнивай символы (руны), а не байты.
// Ожидаемая сложность: O(n) по времени.
func IsPalindrome(s string) bool {
	i, j := 0, len(s) // i: начало непросмотренной части, j: её конец (не включая)
	for i < j {
		l, ls := utf8.DecodeRuneInString(s[i:])
		if !isAlnum(l) {
			i += ls
			continue
		}
		r, rs := utf8.DecodeLastRuneInString(s[:j])
		if !isAlnum(r) {
			j -= rs
			continue
		}
		if unicode.ToLower(l) != unicode.ToLower(r) {
			return false
		}
		i += ls
		j -= rs
	}
	return true
}

func isAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}
