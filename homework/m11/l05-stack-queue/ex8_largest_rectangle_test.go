package l05stackqueue

import (
	"slices"
	"testing"
)

func TestLargestRectangle(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"пустая", nil, 0},
		{"один столбик", []int{7}, 7},
		{"нулевой столбик", []int{0}, 0},
		{"пример", []int{2, 1, 5, 6, 2, 3}, 10},
		{"два", []int{2, 4}, 4},
		{"все одинаковые", []int{3, 3, 3, 3}, 12},
		{"возрастает", []int{1, 2, 3, 4, 5}, 9},
		{"убывает", []int{5, 4, 3, 2, 1}, 9},
		{"низкий во всю ширину", []int{4, 1, 4, 1, 4, 1, 4}, 7},
		{"ноль делит гистограмму", []int{3, 3, 0, 2, 2, 2}, 6},
		{"пик в середине", []int{1, 1, 8, 1, 1}, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LargestRectangle(slices.Clone(tt.in)); got != tt.want {
				t.Errorf("LargestRectangle(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestLargestRectangleBig(t *testing.T) {
	const n = 200_000
	up := make([]int, n)
	for i := range up {
		up[i] = i + 1 // 1, 2, ..., n: лучший прямоугольник высоты n/2 на ширину n/2+1
	}
	tall := make([]int, n)
	for i := range tall {
		tall[i] = 1_000_000_000 // площадь 2*10^14: в int32 не влезет
	}
	var got1, got2 int
	within(t, func() {
		got1 = LargestRectangle(up)
		got2 = LargestRectangle(tall)
	})
	if want := (n / 2) * (n/2 + 1); got1 != want {
		t.Errorf("LargestRectangle(1..%d) = %d, want %d", n, got1, want)
	}
	if want := 1_000_000_000 * n; got2 != want {
		t.Errorf("LargestRectangle(%d столбиков по 10^9) = %d, want %d", n, got2, want)
	}
}
