package l10graphs

import "testing"

func TestFindCircleNum(t *testing.T) {
	tests := []struct {
		name string
		in   [][]int
		want int
	}{
		{"пустой граф", [][]int{}, 0},
		{"один город", [][]int{{1}}, 1},
		{"две провинции", [][]int{{1, 1, 0}, {1, 1, 0}, {0, 0, 1}}, 2},
		{"нет дорог", [][]int{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}, 3},
		// это не сетка: 0 и 2 связаны, хотя «клетки» не соседние
		{"не острова", [][]int{{1, 0, 1}, {0, 1, 0}, {1, 0, 1}}, 2},
		{"связь через посредника", [][]int{
			{1, 0, 0, 1},
			{0, 1, 1, 0},
			{0, 1, 1, 1},
			{1, 0, 1, 1},
		}, 1},
		{"цикл", [][]int{
			{1, 1, 0, 1},
			{1, 1, 1, 0},
			{0, 1, 1, 1},
			{1, 0, 1, 1},
		}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FindCircleNum(tt.in); got != tt.want {
				t.Errorf("FindCircleNum(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestFindCircleNumBig(t *testing.T) {
	const n = 2000
	m := make([][]int, n)
	for i := range m {
		m[i] = make([]int, n)
		m[i][i] = 1
	}
	link := func(a, b int) { m[a][b], m[b][a] = 1, 1 }
	// чётные города: одна длинная цепочка 0-2-4-...; нечётные по парам
	for i := 0; i+2 < n; i += 2 {
		link(i, i+2)
	}
	for i := 1; i+2 < n; i += 4 {
		link(i, i+2)
	}
	want := 1 + n/4
	var got int
	within(t, func() { got = FindCircleNum(m) })
	if got != want {
		t.Errorf("FindCircleNum(%d городов) = %d, want %d", n, got, want)
	}
}
