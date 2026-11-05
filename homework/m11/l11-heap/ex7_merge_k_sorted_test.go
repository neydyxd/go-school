package l11heap

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestMergeKSorted(t *testing.T) {
	tests := []struct {
		name string
		in   [][]int
		want []int
	}{
		{"нет списков", nil, nil},
		{"только пустые", [][]int{nil, {}}, nil},
		{"один список", [][]int{{1, 2, 3}}, []int{1, 2, 3}},
		{"LeetCode", [][]int{{1, 4, 5}, {1, 3, 4}, {2, 6}}, []int{1, 1, 2, 3, 4, 4, 5, 6}},
		{"пустой посередине", [][]int{{2}, {}, {1}}, []int{1, 2}},
		{"повторы и отрицательные", [][]int{{-3, -3, 0}, {-3, 5}, {0, 0}}, []int{-3, -3, -3, 0, 0, 0, 5}},
		{"разной длины", [][]int{{10}, {1, 2, 3, 4, 5, 6}, {7}}, []int{1, 2, 3, 4, 5, 6, 7, 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := fmt.Sprint(tt.in)
			got := MergeKSorted(tt.in)
			if len(got) != len(tt.want) || (len(got) > 0 && !slices.Equal(got, tt.want)) {
				t.Errorf("MergeKSorted(%v) = %v, want %v", before, got, tt.want)
			}
			if after := fmt.Sprint(tt.in); after != before {
				t.Errorf("MergeKSorted изменила вход: было %s, стало %s", before, after)
			}
		})
	}
}

func TestMergeKSortedBig(t *testing.T) {
	const k, m = 50_000, 10 // 50 000 списков по 10 элементов
	r := rand.New(rand.NewPCG(23, 24))
	lists := make([][]int, k)
	var all []int
	for i := range lists {
		l := make([]int, m)
		for j := range l {
			l[j] = r.IntN(1 << 30)
		}
		slices.Sort(l)
		lists[i] = l
		all = append(all, l...)
	}
	slices.Sort(all)
	var got []int
	within(t, func() { got = MergeKSorted(lists) })
	if !slices.Equal(got, all) {
		t.Errorf("MergeKSorted(%d списков по %d): результат неверный (длина %d, want %d)", k, m, len(got), len(all))
	}
}
