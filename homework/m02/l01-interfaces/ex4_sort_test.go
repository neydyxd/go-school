package l01interfaces

import (
	"slices"
	"sort"
	"testing"
)

var _ sort.Interface = ByLenThenAlpha(nil)

func TestByLenThenAlpha(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{"пусто", []string{}, []string{}},
		{"по длине", []string{"ccc", "a", "bb"}, []string{"a", "bb", "ccc"}},
		{"равная длина по алфавиту", []string{"go", "ab", "zz", "a"}, []string{"a", "ab", "go", "zz"}},
		{"руны, а не байты", []string{"lion", "кот", "ox"}, []string{"ox", "кот", "lion"}},
		{"дубликаты", []string{"b", "a", "b"}, []string{"a", "b", "b"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Clone(tt.in)
			sort.Sort(ByLenThenAlpha(got))
			if !slices.Equal(got, tt.want) {
				t.Errorf("sort.Sort(ByLenThenAlpha(%q)) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
