package l06strings

import "unicode"

// IsPalindrome сообщает, читается ли фраза одинаково в обе стороны.
// Учитываются только буквы и цифры (unicode.IsLetter, unicode.IsDigit),
// регистр не важен: «А роза упала на лапу Азора» это палиндром.
// Строка без букв и цифр (в том числе пустая) считается палиндромом.
func IsPalindrome(s string) bool {
	var rs []rune
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			rs = append(rs, unicode.ToLower(r))
		}
	}
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		if rs[i] != rs[j] {
			return false
		}
	}
	return true
}
