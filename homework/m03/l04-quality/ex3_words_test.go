package l04quality

import (
	"maps"
	"strings"
	"testing"
)

func TestCountWords(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want map[string]int
	}{
		{"пусто", "", map[string]int{}},
		{"только разделители", " ,.!\n\t", map[string]int{}},
		{"регистр", "Go, go GO! Гофер", map[string]int{"go": 3, "гофер": 1}},
		{"цифры и строки", "go 1.24\ngo 1.27", map[string]int{"go": 2, "1": 2, "24": 1, "27": 1}},
		{"подчёркивание разделяет", "snake_case", map[string]int{"snake": 1, "case": 1}},
		{"не только арабские цифры", "x² Ⅻ ½ 3", map[string]int{"x²": 1, "ⅻ": 1, "½": 1, "3": 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CountWords(tt.in); !maps.Equal(got, tt.want) {
				t.Errorf("CountWords(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// sampleText это 200 строк текста со словами в разном регистре.
var sampleText = strings.Repeat("Go is fun, and go is fast. Гофер любит Go\n", 200)

func TestCountWordsAllocs(t *testing.T) {
	if raceEnabled {
		t.Skip("под -race AllocsPerRun завышен: детектор сбрасывает sync.Pool в regexp")
	}
	got := CountWords(sampleText)
	if got["go"] != 600 || got["гофер"] != 200 {
		t.Fatalf("CountWords(sampleText): go=%d гофер=%d, want 600 и 200", got["go"], got["гофер"])
	}
	allocs := testing.AllocsPerRun(10, func() { CountWords(sampleText) })
	const limit = 1000
	if allocs > limit {
		t.Errorf("CountWords(200 строк): %.0f аллокаций на вызов, want не больше %d", allocs, limit)
	}
}

func BenchmarkCountWords(b *testing.B) {
	for b.Loop() {
		CountWords(sampleText)
	}
}
