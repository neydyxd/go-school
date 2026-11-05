package l08sorting

import (
	"cmp"
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestMergeSortFunc(t *testing.T) {
	noImports(t, "ex2_merge_sort.go", "sort", "slices")
	tests := []struct {
		name string
		in   []int
	}{
		{"пустой", []int{}},
		{"один", []int{1}},
		{"два наоборот", []int{2, 1}},
		{"пример", []int{38, 27, 43, 3, 9, 82, 10}},
		{"повторы", []int{3, 1, 3, 1, 2}},
		{"нечётная длина", []int{5, 4, 3, 2, 1}},
		{"отрицательные", []int{0, -2, 5, -9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := slices.Clone(tt.in)
			var got []int
			within(t, func() { got = MergeSortFunc(in, cmp.Compare[int]) })
			want := slices.Clone(tt.in)
			slices.Sort(want)
			if !slices.Equal(got, want) {
				t.Errorf("MergeSortFunc(%v) = %v, want %v", tt.in, got, want)
			}
			if !slices.Equal(in, tt.in) {
				t.Errorf("MergeSortFunc(%v) изменил вход: %v", tt.in, in)
			}
		})
	}
}

func TestMergeSortFuncStable(t *testing.T) {
	type rec struct {
		key int
		id  string
	}
	in := []rec{{2, "a"}, {1, "b"}, {2, "c"}, {1, "d"}, {0, "e"}, {2, "f"}, {1, "g"}}
	byKey := func(a, b rec) int { return cmp.Compare(a.key, b.key) }
	var got []rec
	within(t, func() { got = MergeSortFunc(in, byKey) })
	want := []rec{{0, "e"}, {1, "b"}, {1, "d"}, {1, "g"}, {2, "a"}, {2, "c"}, {2, "f"}}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("MergeSortFunc неустойчива: %v, want %v", got, want)
	}
	// по убыванию: функция сравнения задаёт порядок, равные всё так же по входу
	var desc []rec
	within(t, func() { desc = MergeSortFunc(in, func(a, b rec) int { return byKey(b, a) }) })
	wantDesc := []rec{{2, "a"}, {2, "c"}, {2, "f"}, {1, "b"}, {1, "d"}, {1, "g"}, {0, "e"}}
	if fmt.Sprint(desc) != fmt.Sprint(wantDesc) {
		t.Errorf("MergeSortFunc по убыванию: %v, want %v", desc, wantDesc)
	}
}

func TestMergeSortFuncBig(t *testing.T) {
	const n = 1_000_000
	r := rand.New(rand.NewPCG(3, 4))
	in := make([]int, n)
	for i := range in {
		in[i] = r.IntN(1000) // много равных
	}
	var got []int
	within(t, func() { got = MergeSortFunc(in, cmp.Compare[int]) })
	want := slices.Clone(in)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("MergeSortFunc(%d чисел): результат неверный", n)
	}
}
