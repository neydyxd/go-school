package l10graphs

import (
	"strings"
	"testing"
)

func TestNumIslands(t *testing.T) {
	tests := []struct {
		name string
		in   [][]byte
		want int
	}{
		{"пустая сетка", nil, 0},
		{"одна вода", grid("0"), 0},
		{"одна суша", grid("1"), 1},
		{"один большой", grid("11110", "11010", "11000", "00000"), 1},
		{"три острова", grid("11000", "11000", "00100", "00011"), 3},
		{"диагональ не соединяет", grid("101", "010", "101"), 5},
		{"кольцо с озером", grid("111", "101", "111"), 1},
		{"змейка", grid("11111", "00001", "11111", "10000", "11111"), 1},
		{"одна строка", grid("1011101"), 3},
		{"один столбец", grid("1", "0", "1", "1"), 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var show []string
			for _, r := range tt.in {
				show = append(show, string(r))
			}
			if got := NumIslands(tt.in); got != tt.want {
				t.Errorf("NumIslands(%v) = %d, want %d", show, got, tt.want)
			}
		})
	}
}

func TestNumIslandsBig(t *testing.T) {
	const n = 1000
	// сплошная суша: один остров из миллиона клеток
	full := make([][]byte, n)
	for i := range full {
		full[i] = []byte(strings.Repeat("1", n))
	}
	var got int
	within(t, func() { got = NumIslands(full) })
	if got != 1 {
		t.Errorf("NumIslands(сплошная суша %dx%d) = %d, want 1", n, n, got)
	}
	// шахматка: каждая клетка суши сама по себе остров
	chess := make([][]byte, n)
	for i := range chess {
		chess[i] = make([]byte, n)
		for j := range chess[i] {
			chess[i][j] = "01"[(i+j)%2]
		}
	}
	within(t, func() { got = NumIslands(chess) })
	if got != n*n/2 {
		t.Errorf("NumIslands(шахматка %dx%d) = %d, want %d", n, n, got, n*n/2)
	}
}
