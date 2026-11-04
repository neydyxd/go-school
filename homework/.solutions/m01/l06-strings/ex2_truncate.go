package l06strings

import "unicode/utf8"

// Truncate укорачивает строку для превью (заголовок письма, имя в списке).
//
// Если в s не больше n рун, вернуть s как есть.
// Иначе вернуть первые n-1 рун и символ «…» (одна руна U+2026), чтобы
// итог занимал ровно n рун.
// n <= 0: пустая строка. Резать посреди руны нельзя: итог всегда валидный UTF-8.
func Truncate(s string, n int) string {
	if n <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	count := 0
	for i := range s { // i это байтовые позиции начала рун
		if count == n-1 {
			return s[:i] + "…"
		}
		count++
	}
	return s // недостижимо: длина уже проверена
}
