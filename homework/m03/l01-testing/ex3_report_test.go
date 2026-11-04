package l01testing

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRenderReportGolden(t *testing.T) {
	tests := []struct {
		name   string
		items  []Item
		golden string
	}{
		{"обычный чек", []Item{
			{"Яблоки", 3, 9000},
			{"Хлеб", 1, 5450},
			{"Спички", 10, 5},
			{"Сыр", 2, 123456},
		}, "report.golden"},
		{"пустой чек", nil, "empty.golden"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := RenderReport(tt.items)
			path := filepath.Join("testdata", tt.golden)
			want, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("не прочитать golden-файл: %v", err)
			}
			if got != string(want) {
				t.Errorf("RenderReport() не совпал с %s\n--- got:\n%s--- want:\n%s", path, got, want)
			}
		})
	}
}
