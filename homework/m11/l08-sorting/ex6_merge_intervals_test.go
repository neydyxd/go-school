package l08sorting

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestMergeIntervals(t *testing.T) {
	tests := []struct {
		name string
		in   [][2]int
		want [][2]int
	}{
		{"пустой", nil, nil},
		{"один", [][2]int{{1, 3}}, [][2]int{{1, 3}}},
		{"пример", [][2]int{{1, 3}, {2, 6}, {8, 10}, {15, 18}}, [][2]int{{1, 6}, {8, 10}, {15, 18}}},
		{"касаются", [][2]int{{1, 4}, {4, 5}}, [][2]int{{1, 5}}},
		{"не по порядку", [][2]int{{8, 10}, {1, 3}, {2, 6}}, [][2]int{{1, 6}, {8, 10}}},
		{"вложенный", [][2]int{{1, 10}, {2, 3}, {4, 5}}, [][2]int{{1, 10}}},
		{"вложенный не укорачивает", [][2]int{{1, 10}, {2, 3}, {9, 12}}, [][2]int{{1, 12}}},
		{"точки", [][2]int{{5, 5}, {1, 1}, {1, 2}}, [][2]int{{1, 2}, {5, 5}}},
		{"цепочка", [][2]int{{4, 5}, {3, 4}, {2, 3}, {1, 2}}, [][2]int{{1, 5}}},
		{"рядом, но не касаются", [][2]int{{1, 2}, {3, 4}}, [][2]int{{1, 2}, {3, 4}}},
		{"отрицательные", [][2]int{{-5, -1}, {-2, 0}}, [][2]int{{-5, 0}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := slices.Clone(tt.in)
			var got [][2]int
			within(t, func() { got = MergeIntervals(in) })
			if len(got) != len(tt.want) || fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Errorf("MergeIntervals(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestMergeIntervalsBig(t *testing.T) {
	const n = 200_000
	in := make([][2]int, n)
	for i := range in {
		in[i] = [2]int{3 * i, 3*i + 1} // ни один не пересекается
	}
	r := rand.New(rand.NewPCG(9, 10))
	r.Shuffle(n, func(i, j int) { in[i], in[j] = in[j], in[i] })
	var got [][2]int
	within(t, func() { got = MergeIntervals(in) })
	if len(got) != n {
		t.Fatalf("MergeIntervals(%d непересекающихся) вернул %d отрезков", n, len(got))
	}
	for i, iv := range got {
		if iv != [2]int{3 * i, 3*i + 1} {
			t.Fatalf("MergeIntervals(большой вход)[%d] = %v, want %v", i, iv, [2]int{3 * i, 3*i + 1})
		}
	}
}
