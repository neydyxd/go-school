package l01memory

import (
	"strings"
	"testing"
)

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name   string
		in     string
		want   Header
		wantOK bool
	}{
		{"обычный", "Content-Type: text/html", Header{"Content-Type", "text/html"}, true},
		{"без пробела", "Accept:*/*", Header{"Accept", "*/*"}, true},
		{"пробелы и табы по краям", " \tX-Id \t:\t 42  ", Header{"X-Id", "42"}, true},
		{"\\r в конце строки", "Accept: */*\r", Header{"Accept", "*/*"}, true},
		{"двоеточие в значении", "Host: example.com:8080", Header{"Host", "example.com:8080"}, true},
		{"пустое значение", "X-Empty:", Header{"X-Empty", ""}, true},
		{"кириллица", "X-Город: Нижний Новгород", Header{"X-Город", "Нижний Новгород"}, true},
		{"нет двоеточия", "Content-Type text/html", Header{}, false},
		{"пустое имя", ": value", Header{}, false},
		{"имя из пробелов", "   : value", Header{}, false},
		{"пустая строка", "", Header{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ParseHeader(tt.in)
			if ok != tt.wantOK || (ok && got != tt.want) {
				t.Errorf("ParseHeader(%q) = %+v, %v, want %+v, %v", tt.in, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestParseHeaderAllocs(t *testing.T) {
	if raceEnabled {
		t.Skip("под -race счётчик аллокаций не показателен")
	}
	lines := []string{
		"Content-Type: text/html; charset=utf-8",
		"  X-Request-Id :  " + strings.Repeat("ab", 40) + "  ",
		"Host: example.com:8080",
		"broken line",
	}
	for _, line := range lines {
		allocs := testing.AllocsPerRun(100, func() { ParseHeader(line) })
		if allocs != 0 {
			t.Errorf("ParseHeader(%q): %.0f аллокаций на вызов, want 0", line, allocs)
		}
	}
}
