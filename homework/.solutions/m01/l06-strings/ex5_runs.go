package l06strings

import (
	"strconv"
	"strings"
)

// CompressRuns сжимает серии одинаковых символов: символ, затем длина серии,
// если она больше 1. Работает по рунам.
// "aaab" -> "a3b", "пппр" -> "п3р", "abc" -> "abc", "zzzzzzzzzzzz" -> "z12".
// Пустая строка: пустая строка.
func CompressRuns(s string) string {
	var b strings.Builder
	var prev rune
	count := 0
	flush := func() {
		if count == 0 {
			return
		}
		b.WriteRune(prev)
		if count > 1 {
			b.WriteString(strconv.Itoa(count))
		}
	}
	for _, r := range s {
		if count > 0 && r == prev {
			count++
			continue
		}
		flush()
		prev, count = r, 1
	}
	flush()
	return b.String()
}
