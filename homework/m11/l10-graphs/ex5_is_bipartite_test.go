package l10graphs

import "testing"

func TestIsBipartite(t *testing.T) {
	tests := []struct {
		name string
		in   [][]int
		want bool
	}{
		{"пустой граф", [][]int{}, true},
		{"одна вершина", [][]int{{}}, true},
		{"ребро", [][]int{{1}, {0}}, true},
		{"треугольник", [][]int{{1, 2}, {0, 2}, {0, 1}}, false},
		{"квадрат", [][]int{{1, 3}, {0, 2}, {1, 3}, {0, 2}}, true},
		{"LeetCode: нет", [][]int{{1, 2, 3}, {0, 2}, {0, 1, 3}, {0, 2}}, false},
		// вершина 0 изолирована, нечётный цикл в другой компоненте
		{"несвязный с треугольником", [][]int{{}, {2, 3}, {1, 3}, {1, 2}}, false},
		{"несвязный, обе доли", [][]int{{1}, {0}, {3}, {2}, {}}, true},
		{"цикл из 5", [][]int{{1, 4}, {0, 2}, {1, 3}, {2, 4}, {3, 0}}, false},
		{"цикл из 6", [][]int{{1, 5}, {0, 2}, {1, 3}, {2, 4}, {3, 5}, {4, 0}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsBipartite(tt.in); got != tt.want {
				t.Errorf("IsBipartite(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsBipartiteBig(t *testing.T) {
	const n = 200_000
	// чётный цикл на n вершинах, потом в конце отдельный треугольник
	g := make([][]int, n+3)
	for i := range n {
		j := (i + 1) % n
		g[i] = append(g[i], j)
		g[j] = append(g[j], i)
	}
	var got bool
	within(t, func() { got = IsBipartite(g[:n]) })
	if !got {
		t.Errorf("IsBipartite(чётный цикл из %d) = false, want true", n)
	}
	a, b, c := n, n+1, n+2
	g[a], g[b], g[c] = []int{b, c}, []int{a, c}, []int{a, b}
	within(t, func() { got = IsBipartite(g) })
	if got {
		t.Errorf("IsBipartite(цикл из %d + отдельный треугольник) = true, want false", n)
	}
}
