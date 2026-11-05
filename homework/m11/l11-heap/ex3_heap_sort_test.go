package l11heap

import (
	"math"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestHeapSort(t *testing.T) {
	tests := []struct {
		name string
		in   []int
	}{
		{"пустой", nil},
		{"один", []int{42}},
		{"два", []int{2, 1}},
		{"уже отсортирован", []int{1, 2, 3, 4, 5}},
		{"по убыванию", []int{5, 4, 3, 2, 1}},
		{"повторы", []int{3, 1, 2, 3, 1, 2, 3}},
		{"все одинаковые", []int{7, 7, 7, 7}},
		{"края int", []int{0, math.MaxInt, -3, math.MinInt, 3}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := slices.Clone(tt.in)
			HeapSort(a)
			want := slices.Sorted(slices.Values(tt.in))
			if !slices.Equal(a, want) {
				t.Errorf("HeapSort(%v) = %v, want %v", tt.in, a, want)
			}
		})
	}
}

func TestHeapSortBig(t *testing.T) {
	const n = 500_000
	r := rand.New(rand.NewPCG(17, 18))
	a := make([]int, n)
	for i := range a {
		a[i] = r.IntN(2*n) - n
	}
	want := slices.Sorted(slices.Values(a))
	within(t, func() { HeapSort(a) })
	if !slices.Equal(a, want) {
		t.Errorf("HeapSort(%d случайных): результат не отсортирован или потерял элементы", n)
	}
}
