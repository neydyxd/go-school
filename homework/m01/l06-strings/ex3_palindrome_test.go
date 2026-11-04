package l06strings

import "testing"

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"классика", "А роза упала на лапу Азора", true},
		{"латиница с пунктуацией", "A man, a plan, a canal: Panama!", true},
		{"цифры", "12321", true},
		{"цифра ломает палиндром", "аба1", false},
		{"цифры внутри слов", "1а2а1", true},
		{"разные цифры по краям", "1аба2", false},
		{"не палиндром", "привет", false},
		{"знак в конце", "абвба!", true},
		{"разный регистр", "Шалаш", true},
		{"кириллица не по байтам", "аб", false},
		{"одна буква", "я", true},
		{"только знаки", "?!", true},
		{"пустая", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPalindrome(tt.in); got != tt.want {
				t.Errorf("IsPalindrome(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
