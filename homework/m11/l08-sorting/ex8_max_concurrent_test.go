package l08sorting

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestMaxConcurrent(t *testing.T) {
	tests := []struct {
		name string
		in   [][2]int
		want int
	}{
		{"пустой", nil, 0},
		{"одна", [][2]int{{1, 5}}, 1},
		{"стык не пересечение", [][2]int{{1, 5}, {5, 8}}, 1},
		{"пересекаются", [][2]int{{1, 5}, {4, 8}}, 2},
		{"вложенные", [][2]int{{1, 10}, {2, 3}, {4, 5}}, 2},
		{"три в одной точке", [][2]int{{1, 4}, {2, 5}, {3, 6}, {6, 7}}, 3},
		{"одинаковые", [][2]int{{0, 1}, {0, 1}, {0, 1}}, 3},
		{"не по порядку", [][2]int{{7, 9}, {0, 30}, {5, 10}, {15, 20}}, 3},
		{"большие времена", [][2]int{{999_999_998, 1_000_000_000}, {0, 999_999_999}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := slices.Clone(tt.in)
			var got int
			within(t, func() { got = MaxConcurrent(in) })
			if got != tt.want {
				t.Errorf("MaxConcurrent(%v) = %d, want %d", tt.in, got, tt.want)
			}
			if !slices.Equal(in, tt.in) {
				t.Errorf("MaxConcurrent(%v) изменил вход: %v", tt.in, in)
			}
		})
	}
}

func TestMaxConcurrentBig(t *testing.T) {
	const n = 200_000
	r := rand.New(rand.NewPCG(11, 12))
	in := make([][2]int, n)
	for i := range in {
		// цепочки стыкующихся сессий: в каждый момент активна ровно одна
		// из каждой из 1000 «линий»
		k := i / 1000
		in[i] = [2]int{k * 5_000_000, (k + 1) * 5_000_000}
	}
	r.Shuffle(n, func(i, j int) { in[i], in[j] = in[j], in[i] })
	var got int
	within(t, func() { got = MaxConcurrent(in) })
	if got != 1000 {
		t.Fatalf("MaxConcurrent(%d сессий в 1000 линий) = %d, want 1000", n, got)
	}
}
