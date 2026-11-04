package l03hashing

import "testing"

func TestIsIsomorphic(t *testing.T) {
	tests := []struct {
		name string
		s, t string
		want bool
	}{
		{"пустые", "", "", true},
		{"классика", "egg", "add", true},
		{"нет", "foo", "bar", false},
		{"paper title", "paper", "title", true},
		{"нужны оба направления", "badc", "baba", false},
		{"обратное направление", "ab", "aa", false},
		{"разная длина", "ab", "abc", false},
		{"кириллица и латиница", "аб", "ab", true},
		{"кириллица", "абб", "вгг", true},
		{"сам в себя", "abc", "abc", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsIsomorphic(tt.s, tt.t); got != tt.want {
				t.Errorf("IsIsomorphic(%q, %q) = %v, want %v", tt.s, tt.t, got, tt.want)
			}
		})
	}
}

func TestIsIsomorphicBig(t *testing.T) {
	const n = 100_000
	s := make([]rune, n)
	u := make([]rune, n)
	for i := range n {
		s[i] = rune(0x20000 + i)
		u[i] = rune(0x30000 + n - i)
	}
	var got bool
	within(t, func() { got = IsIsomorphic(string(s), string(u)) })
	if !got {
		t.Errorf("IsIsomorphic(большой вход) = false, want true")
	}
}
