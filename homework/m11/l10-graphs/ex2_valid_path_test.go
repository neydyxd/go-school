package l10graphs

import (
	"math/rand/v2"
	"testing"
)

func TestValidPath(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		edges    [][2]int
		src, dst int
		want     bool
	}{
		{"одна вершина", 1, nil, 0, 0, true},
		{"нет рёбер", 2, nil, 0, 1, false},
		{"треугольник", 3, [][2]int{{0, 1}, {1, 2}, {2, 0}}, 0, 2, true},
		{"разные компоненты", 6, [][2]int{{0, 1}, {0, 2}, {3, 5}, {5, 4}, {4, 3}}, 0, 5, false},
		{"ребро в обратную сторону", 2, [][2]int{{1, 0}}, 0, 1, true},
		{"длинная цепочка", 5, [][2]int{{3, 4}, {0, 1}, {2, 3}, {1, 2}}, 0, 4, true},
		{"цикл без цели", 5, [][2]int{{0, 1}, {1, 2}, {2, 0}, {3, 4}}, 1, 4, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidPath(tt.n, tt.edges, tt.src, tt.dst); got != tt.want {
				t.Errorf("ValidPath(%d, %v, %d, %d) = %v, want %v", tt.n, tt.edges, tt.src, tt.dst, got, tt.want)
			}
		})
	}
}

func TestValidPathBig(t *testing.T) {
	const n = 200_000
	// цепочка 0-1-...-(n-2) в перемешанном порядке рёбер, плюс циклы
	// из «обратных» рёбер; вершина n-1 изолирована
	var edges [][2]int
	for i := 0; i+1 < n-1; i++ {
		edges = append(edges, [2]int{i, i + 1})
		if i%3 == 0 && i >= 5 {
			edges = append(edges, [2]int{i, i - 5})
		}
	}
	r := rand.New(rand.NewPCG(5, 6))
	r.Shuffle(len(edges), func(i, j int) { edges[i], edges[j] = edges[j], edges[i] })
	var got1, got2 bool
	within(t, func() {
		got1 = ValidPath(n, edges, 0, n-2)
		got2 = ValidPath(n, edges, 0, n-1)
	})
	if !got1 {
		t.Errorf("ValidPath(цепочка из %d, 0, %d) = false, want true", n-1, n-2)
	}
	if got2 {
		t.Errorf("ValidPath(к изолированной вершине %d) = true, want false", n-1)
	}
}
