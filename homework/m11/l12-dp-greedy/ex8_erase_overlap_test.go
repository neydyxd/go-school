package l12dpgreedy

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestEraseOverlapIntervals(t *testing.T) {
	tests := []struct {
		name string
		in   [][2]int
		want int
	}{
		{"пусто", nil, 0},
		{"один", [][2]int{{1, 2}}, 0},
		{"LeetCode 1", [][2]int{{1, 2}, {2, 3}, {3, 4}, {1, 3}}, 1},
		{"одинаковые", [][2]int{{1, 2}, {1, 2}, {1, 2}}, 2},
		{"касание не пересечение", [][2]int{{1, 2}, {2, 3}}, 0},
		// длинный отрезок начинается раньше всех, но выгодно убрать его
		{"длинный первым", [][2]int{{1, 100}, {11, 22}, {1, 11}, {2, 12}}, 2},
		{"длинный накрывает короткие", [][2]int{{1, 100}, {2, 3}, {3, 4}, {4, 5}}, 1},
		{"отрицательные", [][2]int{{-5, -1}, {-3, 0}, {0, 2}}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := slices.Clone(tt.in)
			if got := EraseOverlapIntervals(in); got != tt.want {
				t.Errorf("EraseOverlapIntervals(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestEraseOverlapIntervalsBig(t *testing.T) {
	// 100 000 непересекающихся отрезков {2i, 2i+1} и 100 000 «длинных»
	// {2i, 2i+3}, каждый пересекает два коротких. Убрать надо все длинные.
	const n = 100_000
	in := make([][2]int, 0, 2*n)
	for i := range n {
		in = append(in, [2]int{2 * i, 2*i + 1}, [2]int{2 * i, 2*i + 3})
	}
	rand.New(rand.NewPCG(27, 28)).Shuffle(len(in), func(i, j int) { in[i], in[j] = in[j], in[i] })
	var got int
	within(t, func() { got = EraseOverlapIntervals(in) })
	if got != n {
		t.Errorf("EraseOverlapIntervals(%d отрезков) = %d, want %d", 2*n, got, n)
	}
}
