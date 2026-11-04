package l05slices

import (
	"slices"
	"testing"
)

func TestFilter(t *testing.T) {
	even := func(x int) bool { return x%2 == 0 }
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"чётные из середины", []int{1, 2, 3, 4, 5, 6}, []int{2, 4, 6}},
		{"все подходят", []int{2, 4}, []int{2, 4}},
		{"ничего не подходит", []int{1, 3, 5}, []int{}},
		{"пустой слайс", []int{}, []int{}},
		{"nil", nil, []int{}},
		{"порядок сохраняется", []int{8, 1, 6, 4}, []int{8, 6, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := slices.Clone(tt.in)
			got := Filter(tt.in, even)
			if len(got) != len(tt.want) || !slices.Equal(got, tt.want) {
				t.Errorf("Filter(%v) = %v, want %v", orig, got, tt.want)
			}
			if !slices.Equal(tt.in, orig) {
				t.Errorf("Filter испортил вход: было %v, стало %v", orig, tt.in)
			}
		})
	}
}

func TestFilterNoAliasing(t *testing.T) {
	in := []int{2, 3, 4}
	got := Filter(in, func(int) bool { return true })
	got[0] = 100
	if in[0] != 2 {
		t.Errorf("результат Filter делит массив со входом: запись в результат изменила вход на %v", in)
	}
}
