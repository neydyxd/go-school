package l05stackqueue

import (
	"strings"
	"testing"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"пустая", "", true},
		{"одна пара", "()", true},
		{"три вида подряд", "()[]{}", true},
		{"вложенные", "{[()()]}", true},
		{"перекрёст", "([)]", false},
		{"не того вида", "(]", false},
		{"лишняя открывающая", "((", false},
		{"лишняя закрывающая", "())", false},
		{"закрывающая первой", ")(", false},
		{"одна открывающая", "[", false},
		{"одна закрывающая", "}", false},
		{"с текстом", "f(a[1], {к: 2})", true},
		{"текст без скобок", "привет, мир", true},
		{"текст и ошибка", "if (x > 0] {", false},
		{"счётчики сошлись, порядок нет", "[(])", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValid(tt.in); got != tt.want {
				t.Errorf("IsValid(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsValidBig(t *testing.T) {
	const n = 300_000
	nested := strings.Repeat("(", n) + strings.Repeat(")", n)
	broken := strings.Repeat("([{", n) + "]" + strings.Repeat("}])", n)
	var got1, got2 bool
	within(t, func() {
		got1 = IsValid(nested)
		got2 = IsValid(broken)
	})
	if !got1 {
		t.Errorf("IsValid(%d вложенных пар) = false, want true", n)
	}
	if got2 {
		t.Errorf("IsValid(длинная строка с лишней ']') = true, want false")
	}
}
