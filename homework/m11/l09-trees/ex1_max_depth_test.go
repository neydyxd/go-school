package l09trees

import "testing"

func TestMaxDepth(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"пустое дерево", "", 0},
		{"один узел", "1", 1},
		{"классика", "3 9 20 # # 15 7", 3},
		{"только левый ребёнок", "1 2", 2},
		{"только правый ребёнок", "1 # 2", 2},
		{"глубокая левая ветка", "1 2 3 4 # # # 5", 4},
		{"зигзаг", "1 2 # # 3 4 # # 5", 5},
		{"полное дерево", "1 2 3 4 5 6 7", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MaxDepth(tree(tt.in)); got != tt.want {
				t.Errorf("MaxDepth(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestMaxDepthBig(t *testing.T) {
	const n = 100_000
	for _, left := range []bool{true, false} {
		root := stick(n, left)
		var got int
		within(t, func() { got = MaxDepth(root) })
		if got != n {
			t.Errorf("MaxDepth(палка из %d узлов, влево=%v) = %d, want %d", n, left, got, n)
		}
	}
	root := balanced(1, 1<<20-1)
	var got int
	within(t, func() { got = MaxDepth(root) })
	if got != 20 {
		t.Errorf("MaxDepth(полное дерево из 2^20-1 узлов) = %d, want 20", got)
	}
}
