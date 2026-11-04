package l02twopointers

import (
	"fmt"
	"slices"
	"testing"
)

// normTriples упорядочивает каждую тройку и сами тройки, чтобы сравнивать
// ответы без учёта порядка.
func normTriples(ts [][]int) [][]int {
	out := make([][]int, 0, len(ts))
	for _, t := range ts {
		c := slices.Clone(t)
		slices.Sort(c)
		out = append(out, c)
	}
	slices.SortFunc(out, slices.Compare)
	return out
}

func TestThreeSum(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want [][]int
	}{
		{"пустой", nil, nil},
		{"меньше трёх", []int{0, 0}, nil},
		{"классика", []int{-1, 0, 1, 2, -1, -4}, [][]int{{-1, -1, 2}, {-1, 0, 1}}},
		{"только нули", []int{0, 0, 0, 0}, [][]int{{0, 0, 0}}},
		{"нет троек", []int{0, 1, 1}, nil},
		{"повторы", []int{-2, 0, 0, 2, 2}, [][]int{{-2, 0, 2}}},
		{"несколько троек", []int{-4, -2, -2, -2, 0, 1, 2, 2, 2, 3, 3, 4, 4, 6, 6}, [][]int{
			{-4, -2, 6}, {-4, 0, 4}, {-4, 1, 3}, {-4, 2, 2}, {-2, -2, 4}, {-2, 0, 2},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := slices.Clone(tt.in)
			got := normTriples(ThreeSum(in))
			want := normTriples(tt.want)
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Errorf("ThreeSum(%v) = %v, want %v", tt.in, got, want)
			}
			for _, tr := range ThreeSum(slices.Clone(tt.in)) {
				if !slices.IsSorted(tr) {
					t.Errorf("ThreeSum(%v): тройка %v не упорядочена", tt.in, tr)
				}
			}
		})
	}
}

func TestThreeSumBig(t *testing.T) {
	// 5000 нечётных чисел: сумма трёх нечётных нечётна, троек нет,
	// но решение O(n^3) успеет перебрать около 2*10^10 троек.
	nums := make([]int, 0, 5000)
	for x := -4999; x <= 4999; x += 2 {
		nums = append(nums, x)
	}
	var got [][]int
	within(t, func() { got = ThreeSum(nums) })
	if len(got) != 0 {
		t.Errorf("ThreeSum(нечётные числа) = %d троек, want 0", len(got))
	}
}
