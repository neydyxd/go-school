package l07maps

import (
	"maps"
	"testing"
)

func TestWordCount(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want map[string]int
	}{
		{"повторы и регистр", "Мама мыла раму, мама!", map[string]int{"мама": 2, "мыла": 1, "раму": 1}},
		{"цифры это часть слова", "go1 go1 go", map[string]int{"go1": 2, "go": 1}},
		{"разделители подряд", "a--b,,  a", map[string]int{"a": 2, "b": 1}},
		{"только разделители", " ,.! ", map[string]int{}},
		{"пустой текст", "", map[string]int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := WordCount(tt.in)
			if got == nil {
				t.Fatalf("WordCount(%q) = nil, want пустую не-nil мапу", tt.in)
			}
			if !maps.Equal(got, tt.want) {
				t.Errorf("WordCount(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}
