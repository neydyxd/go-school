package l01memory

import (
	"strings"
	"testing"
)

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"пустая строка", "", true},
		{"одна буква", "я", true},
		{"латиница", "abba", true},
		{"нечётная длина", "racecar", true},
		{"кириллица", "шалаш", true},
		{"байты не палиндром, руны да", "жж", true},
		{"смесь алфавитов", "aжa", true},
		{"эмодзи", "🙂x🙂", true},
		{"не палиндром", "abc", false},
		{"регистр важен", "Аа", false},
		{"пробелы считаются", "а а", true},
		{"почти палиндром", "шалаж", false},
		{"разные руны с общим байтом", "жз", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPalindrome(tt.in); got != tt.want {
				t.Errorf("IsPalindrome(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsPalindromeLong(t *testing.T) {
	half := strings.Repeat("топот", 100)
	pal := half + "x" + reverseRunes(half)
	if !IsPalindrome(pal) {
		t.Errorf("IsPalindrome(длинный палиндром из %d рун) = false, want true", len([]rune(pal)))
	}
	if IsPalindrome(pal + "!") {
		t.Errorf("IsPalindrome(длинная строка с лишним символом в конце) = true, want false")
	}
}

func TestIsPalindromeAllocs(t *testing.T) {
	if raceEnabled {
		t.Skip("под -race счётчик аллокаций не показателен")
	}
	short := "шалаш"
	long := strings.Repeat("а", 500)
	for _, s := range []string{short, long} {
		allocs := testing.AllocsPerRun(100, func() { IsPalindrome(s) })
		if allocs != 0 {
			t.Errorf("IsPalindrome(строка из %d рун): %.0f аллокаций на вызов, want 0", len([]rune(s)), allocs)
		}
	}
}

func reverseRunes(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
