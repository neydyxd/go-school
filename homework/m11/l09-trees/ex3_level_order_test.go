package l09trees

import (
	"fmt"
	"testing"
)

func TestLevelOrder(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want [][]int
	}{
		{"пустое дерево", "", nil},
		{"один узел", "1", [][]int{{1}}},
		{"классика", "3 9 20 # # 15 7", [][]int{{3}, {9, 20}, {15, 7}}},
		{"палка влево", "3 2 # 1", [][]int{{3}, {2}, {1}}},
		{"дыры на уровне", "1 2 3 4 # # 5", [][]int{{1}, {2, 3}, {4, 5}}},
		{"полное дерево", "1 2 3 4 5 6 7", [][]int{{1}, {2, 3}, {4, 5, 6, 7}}},
		{"отрицательные и повторы", "0 -1 -1 # 2 2", [][]int{{0}, {-1, -1}, {2, 2}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LevelOrder(tree(tt.in))
			if fmt.Sprint(got) != fmt.Sprint(tt.want) || len(got) != len(tt.want) {
				t.Errorf("LevelOrder(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestLevelOrderBig(t *testing.T) {
	const n = 100_000
	root := stick(n, false) // 1, 2, ..., n вправо
	var got [][]int
	within(t, func() { got = LevelOrder(root) })
	if len(got) != n {
		t.Fatalf("LevelOrder(палка из %d) вернула %d уровней, want %d", n, len(got), n)
	}
	for _, i := range []int{0, 1, n / 2, n - 1} {
		if len(got[i]) != 1 || got[i][0] != i+1 {
			t.Errorf("LevelOrder(палка из %d): уровень %d = %v, want [%d]", n, i, got[i], i+1)
		}
	}
	wide := balanced(1, 1<<18-1)
	within(t, func() { got = LevelOrder(wide) })
	if len(got) != 18 || len(got[17]) != 1<<17 {
		t.Errorf("LevelOrder(полное дерево 2^18-1): %d уровней, want 18", len(got))
	}
}
