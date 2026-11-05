package l01memory

import "unicode/utf8"

// IsPalindrome идёт двумя индексами навстречу и декодирует по одной руне
// с каждого конца. Строка не копируется, слайсов нет: ноль аллокаций.
func IsPalindrome(s string) bool {
	i, j := 0, len(s)
	for i < j {
		l, ls := utf8.DecodeRuneInString(s[i:j])
		r, rs := utf8.DecodeLastRuneInString(s[i:j])
		if l != r {
			return false
		}
		i += ls
		j -= rs
	}
	return true
}
