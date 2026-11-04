package l04slidingwindow

import (
	"strings"
	"testing"
)

func TestCharacterReplacement(t *testing.T) {
	tests := []struct {
		name string
		s    string
		k    int
		want int
	}{
		{"пустая", "", 2, 0},
		{"классика", "ABAB", 2, 4},
		{"с одной заменой", "AABABBA", 1, 4},
		{"без замен", "AABBBA", 0, 3},
		{"k больше длины", "ABC", 10, 3},
		{"кириллица в рунах", "ааббба", 1, 4},
		{"один символ", "Я", 0, 1},
		{"всё одинаковое", "ZZZZ", 1, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CharacterReplacement(tt.s, tt.k); got != tt.want {
				t.Errorf("CharacterReplacement(%q, %d) = %d, want %d", tt.s, tt.k, got, tt.want)
			}
		})
	}
}

func TestCharacterReplacementBig(t *testing.T) {
	const n = 200_000
	s := strings.Repeat("A", n/2) + strings.Repeat("B", n/2)
	var got int
	within(t, func() { got = CharacterReplacement(s, 1000) })
	if want := n/2 + 1000; got != want {
		t.Errorf("CharacterReplacement(большой вход, 1000) = %d, want %d", got, want)
	}
}
