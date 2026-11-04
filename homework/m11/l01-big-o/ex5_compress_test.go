package l01bigo

import (
	"strings"
	"testing"
)

func TestCompress(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"пустая", "", ""},
		{"один символ", "x", "x"},
		{"без повторов", "abc", "abc"},
		{"латиница", "aabcccccaaa", "a2bc5a3"},
		{"кириллица", "ааабвв", "а3бв2"},
		{"двузначная длина", "aaaaaaaaaaaa", "a12"},
		{"эмодзи", "🙂🙂🙂x", "🙂3x"},
		{"пробелы тоже символы", "a  b", "a 2b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compress(tt.in); got != tt.want {
				t.Errorf("Compress(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCompressBig(t *testing.T) {
	const n = 400_000
	in := strings.Repeat("аб", n/2) // ни одной серии длиннее 1
	var got string
	within(t, func() { got = Compress(in) })
	if got != in {
		t.Errorf("Compress(\"абаб...\" из %d рун) вернула строку длиной %d байт, want %d", n, len(got), len(in))
	}
}
