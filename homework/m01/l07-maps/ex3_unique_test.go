package l07maps

import (
	"slices"
	"testing"
)

func TestUnique(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"повторы вразброс", []string{"b", "a", "b", "c", "a"}, []string{"b", "a", "c"}},
		{"без повторов", []string{"z", "y"}, []string{"z", "y"}},
		{"все одинаковые", []string{"x", "x", "x"}, []string{"x"}},
		{"пустая строка тоже значение", []string{"", "a", ""}, []string{"", "a"}},
		{"кириллица", []string{"ёж", "еж", "ёж"}, []string{"ёж", "еж"}},
		{"пустой", []string{}, []string{}},
		{"nil", nil, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := slices.Clone(tt.in)
			got := Unique(tt.in)
			if len(got) != len(tt.want) || !slices.Equal(got, tt.want) {
				t.Errorf("Unique(%q) = %q, want %q", orig, got, tt.want)
			}
			if !slices.Equal(tt.in, orig) {
				t.Errorf("Unique испортил вход: было %q, стало %q", orig, tt.in)
			}
		})
	}
}
