package l12dpgreedy

import (
	"strings"
	"testing"
)

func TestLongestCommonSubsequence(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"пустые", "", "", 0},
		{"одна пустая", "abc", "", 0},
		{"LeetCode 1", "abcde", "ace", 3},
		{"одинаковые", "abc", "abc", 3},
		{"нет общих", "abc", "def", 0},
		{"кириллица", "собака", "кабан", 2},
		// в байтах у «а» и «б» общий первый байт d0
		{"байты не символы", "аб", "ба", 1},
		{"порядок важен", "ab", "ba", 1},
		{"повторы", "aaaa", "aa", 2},
		{"эмодзи", "a🙂b🙂", "🙂🙂", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LongestCommonSubsequence(tt.a, tt.b); got != tt.want {
				t.Errorf("LongestCommonSubsequence(%q, %q) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestLongestCommonSubsequenceBig(t *testing.T) {
	// по 2000 рун в каждой строке, ответ посчитан отдельно
	a := strings.Repeat("абвг", 500)
	b := strings.Repeat("бадг", 500)
	var got int
	within(t, func() { got = LongestCommonSubsequence(a, b) })
	if want := 1000; got != want {
		t.Errorf("LongestCommonSubsequence(\"абвг\"*500, \"бадг\"*500) = %d, want %d", got, want)
	}
}
