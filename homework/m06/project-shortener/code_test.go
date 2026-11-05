package shortener

import (
	"errors"
	"strings"
	"testing"
)

func TestNewCode(t *testing.T) {
	for _, n := range []int{-1, 0} {
		if got := NewCode(n); got != "" {
			t.Errorf("NewCode(%d) = %q, want \"\"", n, got)
		}
	}
	seen := map[string]bool{}
	chars := map[rune]bool{}
	for range 2000 {
		c := NewCode(7)
		if len(c) != 7 {
			t.Fatalf("NewCode(7) = %q, длина %d, want 7", c, len(c))
		}
		for _, r := range c {
			if !strings.ContainsRune(Alphabet, r) {
				t.Fatalf("NewCode(7) = %q: символ %q не из Alphabet", c, r)
			}
			chars[r] = true
		}
		if seen[c] {
			t.Fatalf("NewCode(7) повторил код %q за 2000 вызовов: это не случайность", c)
		}
		seen[c] = true
	}
	if len(chars) < 55 {
		t.Errorf("за 2000 кодов встретилось %d разных символов из 62: распределение перекошено", len(chars))
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name string
		in   string
		ok   bool
	}{
		{"https", "https://go.dev/doc/", true},
		{"http с портом и запросом", "http://localhost:8080/a?b=1#c", true},
		{"кириллический домен", "https://пример.рф/путь", true},
		{"схема в верхнем регистре", "HTTPS://go.dev", true},
		{"ровно 2048 байт", "https://a.ru/" + strings.Repeat("x", 2048-len("https://a.ru/")), true},
		{"2049 байт", "https://a.ru/" + strings.Repeat("x", 2049-len("https://a.ru/")), false},
		{"пусто", "", false},
		{"без схемы", "go.dev/doc", false},
		{"javascript", "javascript:alert(1)", false},
		{"ftp", "ftp://files.example.com/a", false},
		{"file", "file:///etc/passwd", false},
		{"без хоста", "https:///path", false},
		{"userinfo", "https://bank.ru@evil.example/login", false},
		{"битый процент", "https://a.ru/%zz", false},
		{"пробел в хосте", "https://a b.ru", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateURL(tt.in)
			if tt.ok && err != nil {
				t.Errorf("ValidateURL(%q) = %v, want nil", tt.in, err)
			}
			if !tt.ok && !errors.Is(err, ErrInvalidURL) {
				t.Errorf("ValidateURL(%q) = %v, want ошибку с ErrInvalidURL", tt.in, err)
			}
		})
	}
}
