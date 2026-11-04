package l06strings

import (
	"testing"
	"unicode/utf8"
)

func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"латиница", "hello", "olleh"},
		{"кириллица", "привет", "тевирп"},
		{"эмодзи 4 байта", "Go🙂", "🙂oG"},
		{"смешанная", "abв", "вba"},
		{"одна руна", "ж", "ж"},
		{"пустая", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reverse(tt.in)
			if got != tt.want {
				t.Errorf("Reverse(%q) = %q, want %q", tt.in, got, tt.want)
			}
			if !utf8.ValidString(got) {
				t.Errorf("Reverse(%q) вернул невалидный UTF-8: переворачиваешь байты, а не руны", tt.in)
			}
		})
	}
}
