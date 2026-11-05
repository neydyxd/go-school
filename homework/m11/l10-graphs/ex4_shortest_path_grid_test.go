package l10graphs

import "testing"

func TestShortestPathBinaryMatrix(t *testing.T) {
	tests := []struct {
		name string
		in   [][]int
		want int
	}{
		{"одна клетка", [][]int{{0}}, 1},
		{"одна стена", [][]int{{1}}, -1},
		{"по диагонали", [][]int{{0, 1}, {1, 0}}, 2},
		{"обход стены", [][]int{{0, 0, 0}, {1, 1, 0}, {1, 1, 0}}, 4},
		{"старт в стене", [][]int{{1, 0, 0}, {1, 1, 0}, {1, 1, 0}}, -1},
		{"финиш в стене", [][]int{{0, 0}, {0, 1}}, -1},
		{"стена поперёк", [][]int{{0, 1, 0}, {1, 1, 0}, {0, 0, 0}}, -1},
		{"открытое поле", [][]int{{0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}, {0, 0, 0, 0}}, 4},
		// длинный коридор короче, чем кажется: важен первый найденный BFS путь
		{"змейка", [][]int{
			{0, 1, 0, 0, 0},
			{0, 1, 0, 1, 0},
			{0, 1, 0, 1, 0},
			{0, 1, 0, 1, 0},
			{0, 0, 0, 1, 0},
		}, 13},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShortestPathBinaryMatrix(tt.in); got != tt.want {
				t.Errorf("ShortestPathBinaryMatrix(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestShortestPathBinaryMatrixBig(t *testing.T) {
	const n = 1000
	open := make([][]int, n)
	for i := range open {
		open[i] = make([]int, n)
	}
	var got int
	within(t, func() { got = ShortestPathBinaryMatrix(open) })
	if got != n {
		t.Errorf("ShortestPathBinaryMatrix(пустое поле %dx%d) = %d, want %d", n, n, got, n)
	}
	// гребёнка 999x999: в каждой нечётной строке стена с одним проходом,
	// попеременно у правого и левого края; путь петляет по всему полю.
	// Между соседними проходами n-1 шаг (диагональ экономит ход по строке),
	// проходов 499, плюс последний шаг вниз и стартовая клетка.
	const m = 999
	comb := make([][]int, m)
	for i := range comb {
		comb[i] = make([]int, m)
		if i%2 == 1 {
			for j := range comb[i] {
				comb[i][j] = 1
			}
			if i%4 == 1 {
				comb[i][m-1] = 0
			} else {
				comb[i][0] = 0
			}
		}
	}
	want := 499*(m-1) + 1 + 1
	within(t, func() { got = ShortestPathBinaryMatrix(comb) })
	if got != want {
		t.Errorf("ShortestPathBinaryMatrix(гребёнка %dx%d) = %d, want %d", m, m, got, want)
	}
}
