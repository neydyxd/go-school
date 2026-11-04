package l04slidingwindow

import (
	"slices"
	"strings"
	"testing"
)

func TestFindAnagrams(t *testing.T) {
	tests := []struct {
		name string
		s, p string
		want []int
	}{
		{"пустые", "", "", nil},
		{"пустой шаблон", "abc", "", nil},
		{"шаблон длиннее", "ab", "abc", nil},
		{"классика", "cbaebabacd", "abc", []int{0, 6}},
		{"перекрываются", "abab", "ab", []int{0, 1, 2}},
		{"кириллица: позиции в рунах", "кот ток окт", "кто", []int{0, 4, 8}},
		{"нет совпадений", "aaaa", "ab", nil},
		{"вся строка", "listen", "silent", []int{0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindAnagrams(tt.s, tt.p)
			if len(got) != len(tt.want) || (len(got) > 0 && !slices.Equal(got, tt.want)) {
				t.Errorf("FindAnagrams(%q, %q) = %v, want %v", tt.s, tt.p, got, tt.want)
			}
		})
	}
}

func TestFindAnagramsBig(t *testing.T) {
	const n = 200_000
	s := strings.Repeat("ab", n/2)
	p := strings.Repeat("ba", n/4) // длина n/2: подходит каждое окно
	var got []int
	within(t, func() { got = FindAnagrams(s, p) })
	if want := n/2 + 1; len(got) != want {
		t.Fatalf("FindAnagrams(большой вход): найдено %d позиций, want %d", len(got), want)
	}
	for i, v := range got {
		if v != i {
			t.Fatalf("FindAnagrams(большой вход): позиция %d = %d, want %d", i, v, i)
		}
	}
}
