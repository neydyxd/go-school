package l01bigo

import (
	"strconv"
	"strings"
)

// Compress сжимает строку кодированием длин серий (RLE) по символам (рунам):
// серия из одного символа записывается как сам символ, серия длиной k >= 2
// записывается как символ и число k.
//
//	"ааабвв"   -> "а3бв2"
//	"abc"      -> "abc"
//	"aaaaaaaaaaaa" -> "a12"
//
// Пустая строка остаётся пустой. Цифр во входе нет.
// Строка может быть длинной: собирай результат за O(n), а не за O(n^2).
func Compress(s string) string {
	rs := []rune(s)
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(rs); {
		j := i
		for j < len(rs) && rs[j] == rs[i] {
			j++
		}
		b.WriteRune(rs[i])
		if j-i > 1 {
			b.WriteString(strconv.Itoa(j - i))
		}
		i = j
	}
	return b.String()
}
