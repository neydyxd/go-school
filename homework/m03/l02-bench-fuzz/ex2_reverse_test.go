package l02benchfuzz

import (
	"errors"
	"testing"
	"unicode/utf8"
)

func TestReverse(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{"пустая", "", "", nil},
		{"латиница", "Hello", "olleH", nil},
		{"кириллица", "Привет", "тевирП", nil},
		{"иероглифы и эмодзи", "Go语言🙂", "🙂言语oG", nil},
		{"невалидный байт", "\xff", "", ErrInvalidUTF8},
		{"обрезанная руна", "При\xd0", "", ErrInvalidUTF8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Reverse(tt.in)
			if got != tt.want || !errors.Is(err, tt.wantErr) {
				t.Errorf("Reverse(%q) = %q, %v, want %q, %v", tt.in, got, err, tt.want, tt.wantErr)
			}
		})
	}
}

// Запуск фаззера: go test -run=^$ -fuzz=FuzzReverse -fuzztime=20s ./m03/l02-bench-fuzz/
func FuzzReverse(f *testing.F) {
	for _, seed := range []string{"", "a", "Hello, мир", "🙂x", "\xff", "a\xc3"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, s string) {
		rev, err := Reverse(s)
		if !utf8.ValidString(s) {
			if !errors.Is(err, ErrInvalidUTF8) {
				t.Fatalf("Reverse(%q) = %q, %v, want ошибку ErrInvalidUTF8", s, rev, err)
			}
			return
		}
		if err != nil {
			t.Fatalf("Reverse(%q) вернул ошибку %v на валидной строке", s, err)
		}
		if !utf8.ValidString(rev) {
			t.Fatalf("Reverse(%q) = %q: результат невалидный UTF-8", s, rev)
		}
		back, err := Reverse(rev)
		if err != nil || back != s {
			t.Fatalf("Reverse(Reverse(%q)) = %q, %v, want исходную строку", s, back, err)
		}
	})
}
