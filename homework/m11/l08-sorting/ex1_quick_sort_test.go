package l08sorting

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestQuickSort(t *testing.T) {
	noImports(t, "ex1_quick_sort.go", "sort", "slices")
	tests := []struct {
		name string
		in   []int
	}{
		{"пустой", []int{}},
		{"nil", nil},
		{"один", []int{1}},
		{"два по порядку", []int{1, 2}},
		{"два наоборот", []int{2, 1}},
		{"пример", []int{5, 2, 3, 1}},
		{"повторы", []int{5, 1, 1, 2, 0, 0}},
		{"все равны", []int{7, 7, 7, 7}},
		{"отрицательные", []int{3, -1, -7, 0, 2, -1}},
		{"уже отсортирован", []int{1, 2, 3, 4, 5, 6}},
		{"развёрнут", []int{6, 5, 4, 3, 2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Clone(tt.in)
			within(t, func() { QuickSort(got) })
			want := slices.Clone(tt.in)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("QuickSort(%v) = %v, want %v", tt.in, got, want)
			}
		})
	}
}

func TestQuickSortBig(t *testing.T) {
	const n = 1_000_000
	r := rand.New(rand.NewPCG(1, 2))
	random := make([]int, n)
	for i := range random {
		random[i] = r.IntN(1_000_000_000) - 500_000_000
	}
	sorted := make([]int, n)
	for i := range sorted {
		sorted[i] = i
	}
	reversed := make([]int, n)
	for i := range reversed {
		reversed[i] = n - i
	}
	equal := make([]int, n) // все нули
	few := make([]int, n)   // всего три разных значения
	for i := range few {
		few[i] = r.IntN(3)
	}
	cases := []struct {
		name string
		in   []int
	}{
		{"случайный", random},
		{"уже отсортирован", sorted},
		{"развёрнут", reversed},
		{"все равны", equal},
		{"три значения", few},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := slices.Clone(c.in)
			within(t, func() { QuickSort(got) })
			if !slices.IsSorted(got) {
				t.Fatalf("QuickSort(%s, %d чисел): результат не отсортирован", c.name, n)
			}
			want := slices.Clone(c.in)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Fatalf("QuickSort(%s, %d чисел): набор чисел изменился", c.name, n)
			}
		})
	}
}
