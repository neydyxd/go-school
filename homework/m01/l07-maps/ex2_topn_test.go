package l07maps

import (
	"slices"
	"testing"
)

func TestTopN(t *testing.T) {
	freq := map[string]int{"go": 5, "rust": 2, "java": 5, "c": 1, "zig": 2, "odin": 2}
	tests := []struct {
		name string
		freq map[string]int
		n    int
		want []string
	}{
		{"топ-3 с ничьей", freq, 3, []string{"go", "java", "odin"}},
		{"все слова", freq, 6, []string{"go", "java", "odin", "rust", "zig", "c"}},
		{"n больше числа слов", freq, 100, []string{"go", "java", "odin", "rust", "zig", "c"}},
		{"n ноль", freq, 0, []string{}},
		{"n отрицательный", freq, -1, []string{}},
		{"пустая мапа", map[string]int{}, 3, []string{}},
		{"nil мапа", nil, 3, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Повторяем: порядок обхода мапы случаен, ответ не должен от него зависеть.
			for range 50 {
				got := TopN(tt.freq, tt.n)
				if len(got) != len(tt.want) || !slices.Equal(got, tt.want) {
					t.Fatalf("TopN(%v, %d) = %v, want %v", tt.freq, tt.n, got, tt.want)
				}
			}
		})
	}
}
