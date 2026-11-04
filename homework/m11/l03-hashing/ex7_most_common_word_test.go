package l03hashing

import (
	"strings"
	"testing"
)

func TestMostCommonWord(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		banned []string
		want   string
	}{
		{"пустой", "", nil, ""},
		{"классика", "Bob hit a ball, the hit BALL flew far after it was hit.", []string{"hit"}, "ball"},
		{"кириллица и регистр", "Мама мыла раму, МАМА!", nil, "мама"},
		{"знаки без пробелов", "a,b,b;c", nil, "b"},
		{"ничья: меньшее слово", "дом кот кот дом", nil, "дом"},
		{"ничья: меньшее слово позже", "кот дом дом кот", nil, "дом"},
		{"ничья из шести слов", "е д г в б а а б в г д е", nil, "а"},
		{"ничья из восьми слов", "h g f e d c b a a b c d e f g h", nil, "a"},
		{"всё запрещено", "go go", []string{"go"}, ""},
		{"цифры разделяют", "go1go2go", nil, "go"},
		{"нет букв", "123 !!", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MostCommonWord(tt.text, tt.banned); got != tt.want {
				t.Errorf("MostCommonWord(%q, %q) = %q, want %q", tt.text, tt.banned, got, tt.want)
			}
		})
	}
}

func TestMostCommonWordBig(t *testing.T) {
	const n = 100_000
	words := make([]string, 0, n)
	for i := range n {
		switch {
		case i%3 == 0:
			words = append(words, "Гофер")
		case i%2 == 0:
			words = append(words, "спам")
		default:
			words = append(words, "w"+strings.Repeat("x", i%50))
		}
	}
	text := strings.Join(words, ", ")
	var got string
	within(t, func() { got = MostCommonWord(text, []string{"спам"}) })
	if got != "гофер" {
		t.Errorf("MostCommonWord(большой текст) = %q, want %q", got, "гофер")
	}
}
