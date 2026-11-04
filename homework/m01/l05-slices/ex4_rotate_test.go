package l05slices

import (
	"slices"
	"testing"
)

func TestRotateLeft(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		k    int
		want []int
	}{
		{"сдвиг на 2", []int{1, 2, 3, 4, 5}, 2, []int{3, 4, 5, 1, 2}},
		{"сдвиг на 0", []int{1, 2, 3}, 0, []int{1, 2, 3}},
		{"сдвиг на длину", []int{1, 2, 3}, 3, []int{1, 2, 3}},
		{"k больше длины", []int{1, 2, 3}, 7, []int{2, 3, 1}},
		{"отрицательный k", []int{1, 2, 3, 4, 5}, -1, []int{5, 1, 2, 3, 4}},
		{"большой отрицательный k", []int{1, 2, 3}, -8, []int{2, 3, 1}},
		{"один элемент", []int{9}, 5, []int{9}},
		{"пустой", []int{}, 3, []int{}},
		{"nil", nil, 3, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := slices.Clone(tt.in)
			RotateLeft(s, tt.k)
			if !slices.Equal(s, tt.want) {
				t.Errorf("RotateLeft(%v, %d) = %v, want %v", tt.in, tt.k, s, tt.want)
			}
		})
	}
}

func TestRotateLeftNoAllocs(t *testing.T) {
	s := make([]int, 1000)
	for i := range s {
		s[i] = i
	}
	allocs := testing.AllocsPerRun(10, func() { RotateLeft(s, 333) })
	if allocs != 0 {
		t.Errorf("RotateLeft выделил память %.0f раз за вызов, want 0: переставляй на месте", allocs)
	}
}
