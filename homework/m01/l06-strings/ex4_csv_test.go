package l06strings

import (
	"strings"
	"testing"
)

func TestCSVLine(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want string
	}{
		{"простые поля", []string{"a", "b", "c"}, "a,b,c"},
		{"запятая в поле", []string{"a", "b,c"}, `a,"b,c"`},
		{"кавычки в поле", []string{`say "hi"`}, `"say ""hi"""`},
		{"перевод строки", []string{"x\ny"}, "\"x\ny\""},
		{"кириллица", []string{"чай", "1,5 кг"}, `чай,"1,5 кг"`},
		{"пустые поля", []string{"", ""}, ","},
		{"одно поле", []string{"solo"}, "solo"},
		{"нет полей", nil, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CSVLine(tt.in); got != tt.want {
				t.Errorf("CSVLine(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestCSVLineAllocs(t *testing.T) {
	fields := make([]string, 50)
	for i := range fields {
		fields[i] = strings.Repeat("поле,", i%4) + `"x"`
	}
	allocs := testing.AllocsPerRun(20, func() { _ = CSVLine(fields) })
	if allocs > 1 {
		t.Errorf("CSVLine делает %.0f аллокаций, want 1: посчитай длину и вызови Grow", allocs)
	}
}
