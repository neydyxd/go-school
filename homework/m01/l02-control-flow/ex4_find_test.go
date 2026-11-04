package l02controlflow

import "testing"

func TestFind(t *testing.T) {
	tests := []struct {
		name             string
		grid             [][]int
		target           int
		wantRow, wantCol int
		wantOK           bool
	}{
		{"nil-таблица", nil, 1, -1, -1, false},
		{"пустые строки", [][]int{{}, {}}, 1, -1, -1, false},
		{"в первой ячейке", [][]int{{7, 1}, {2, 3}}, 7, 0, 0, true},
		{"в середине", [][]int{{1, 2, 3}, {4, 5, 6}}, 5, 1, 1, true},
		{"нет такого", [][]int{{1, 2}, {3, 4}}, 9, -1, -1, false},
		{"первое из нескольких в строке", [][]int{{0, 5, 5}}, 5, 0, 1, true},
		{"первое из нескольких строк", [][]int{{0, 0}, {0, 5}, {5, 0}}, 5, 1, 1, true},
		{"строки разной длины", [][]int{{1}, {}, {2, 3, 4}}, 4, 2, 2, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, c, ok := Find(tt.grid, tt.target)
			if r != tt.wantRow || c != tt.wantCol || ok != tt.wantOK {
				t.Errorf("Find(%v, %d) = (%d, %d, %v), want (%d, %d, %v)",
					tt.grid, tt.target, r, c, ok, tt.wantRow, tt.wantCol, tt.wantOK)
			}
		})
	}
}
