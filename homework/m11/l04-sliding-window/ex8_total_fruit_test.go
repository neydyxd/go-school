package l04slidingwindow

import "testing"

func TestTotalFruit(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int
	}{
		{"пустой", nil, 0},
		{"одно дерево", []int{5}, 1},
		{"всё влезает", []int{1, 2, 1}, 3},
		{"начать не с первого", []int{0, 1, 2, 2}, 3},
		{"середина", []int{1, 2, 3, 2, 2}, 4},
		{"длинная", []int{3, 3, 3, 1, 2, 1, 1, 2, 3, 3, 4}, 5},
		{"один сорт", []int{7, 7, 7}, 3},
		{"сужение до последнего сорта", []int{1, 2, 1, 3, 3, 3}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TotalFruit(tt.in); got != tt.want {
				t.Errorf("TotalFruit(%v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestTotalFruitBig(t *testing.T) {
	const n = 200_000
	fruits := make([]int, n)
	for i := range fruits {
		fruits[i] = i % 2
	}
	fruits[0] = 2 // первое дерево третьего сорта
	var got int
	within(t, func() { got = TotalFruit(fruits) })
	if got != n-1 {
		t.Errorf("TotalFruit(большой вход) = %d, want %d", got, n-1)
	}
}
