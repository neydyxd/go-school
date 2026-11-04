package l06strings

import (
	"testing"
	"unicode/utf8"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name string
		in   string
		n    int
		want string
	}{
		{"короче лимита", "Go", 5, "Go"},
		{"ровно лимит", "Привет", 6, "Привет"},
		{"длиннее, кириллица", "Привет, мир", 7, "Привет…"},
		{"длиннее, латиница", "hello world", 5, "hell…"},
		{"эмодзи", "🙂🙂🙂🙂", 3, "🙂🙂…"},
		{"лимит 1", "абв", 1, "…"},
		{"лимит 0", "абв", 0, ""},
		{"отрицательный лимит", "абв", -2, ""},
		{"пустая строка", "", 3, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Truncate(tt.in, tt.n)
			if got != tt.want {
				t.Errorf("Truncate(%q, %d) = %q, want %q", tt.in, tt.n, got, tt.want)
			}
			if !utf8.ValidString(got) {
				t.Errorf("Truncate(%q, %d) разрезал руну: %q", tt.in, tt.n, got)
			}
		})
	}
}
