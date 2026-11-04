package l08structs

import "testing"

func TestVisited(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int
	}{
		{"пустой путь", "", 1},
		{"прямая", "RRR", 4},
		{"туда и обратно", "RL", 2},
		{"квадрат", "URDL", 4},
		{"восьмёрка возвращается в старт", "URDLDRUL", 6},
		{"мусор игнорируется", "R?x R", 3},
		{"много повторов", "UDUDUDUD", 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Visited(tt.path); got != tt.want {
				t.Errorf("Visited(%q) = %d, want %d", tt.path, got, tt.want)
			}
		})
	}
}
