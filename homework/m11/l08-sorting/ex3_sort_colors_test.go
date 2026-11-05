package l08sorting

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestSortColors(t *testing.T) {
	noImports(t, "ex3_sort_colors.go", "sort", "slices")
	tests := []struct {
		name string
		in   []int
	}{
		{"пустой", []int{}},
		{"один", []int{2}},
		{"пример", []int{2, 0, 2, 1, 1, 0}},
		{"три", []int{2, 0, 1}},
		{"только двойки", []int{2, 2, 2}},
		{"без единиц", []int{2, 0, 2, 0}},
		{"двойка и ноль на краях", []int{2, 1, 1, 0}},
		{"уже отсортирован", []int{0, 0, 1, 2, 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Clone(tt.in)
			within(t, func() { SortColors(got) })
			want := slices.Clone(tt.in)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("SortColors(%v) = %v, want %v", tt.in, got, want)
			}
		})
	}
}

func TestSortColorsBig(t *testing.T) {
	const n = 1_000_000
	r := rand.New(rand.NewPCG(5, 6))
	in := make([]int, n)
	for i := range in {
		in[i] = r.IntN(3)
	}
	got := slices.Clone(in)
	within(t, func() { SortColors(got) })
	want := slices.Clone(in)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("SortColors(%d чисел): результат неверный", n)
	}
}
