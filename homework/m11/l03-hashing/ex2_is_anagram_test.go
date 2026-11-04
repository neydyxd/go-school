package l03hashing

import (
	"strings"
	"testing"
)

func TestIsAnagram(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want bool
	}{
		{"пустые", "", "", true},
		{"классика", "anagram", "nagaram", true},
		{"не анаграмма", "rat", "car", false},
		{"разная длина", "ab", "abb", false},
		{"разные количества", "aab", "abb", false},
		{"кириллица", "апельсин", "спаниель", true},
		{"регистр важен", "Кот", "ток", false},
		{"те же байты, другие руны", "аё", "БѰ", false},
		{"эмодзи", "🙂🙃", "🙃🙂", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAnagram(tt.a, tt.b); got != tt.want {
				t.Errorf("IsAnagram(%q, %q) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestIsAnagramBig(t *testing.T) {
	a := strings.Repeat("абвгд", 40_000)
	b := strings.Repeat("дгвба", 40_000)
	var got bool
	within(t, func() { got = IsAnagram(a, b) })
	if !got {
		t.Errorf("IsAnagram(большие анаграммы) = false, want true")
	}
}
