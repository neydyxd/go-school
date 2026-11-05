package l09trees

import "testing"

func TestDiameter(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"пустое дерево", "", 0},
		{"один узел", "1", 0},
		{"два узла", "1 2", 1},
		{"классика", "1 2 3 4 5", 3},
		{"палка", "1 2 # 3 # 4", 3},
		// самый длинный путь 8-6-4-2-5-7-9 целиком в левом поддереве
		{"не через корень", "1 2 3 4 5 # # 6 # # 7 8 # # 9", 6},
		{"полное дерево", "1 2 3 4 5 6 7", 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Diameter(tree(tt.in)); got != tt.want {
				t.Errorf("Diameter(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestDiameterBig(t *testing.T) {
	const n = 100_000
	root := stick(n, true)
	var got int
	within(t, func() { got = Diameter(root) })
	if got != n-1 {
		t.Errorf("Diameter(палка из %d) = %d, want %d", n, got, n-1)
	}
	// «вилка»: две палки по n/2 узлов слева и справа от корня
	fork := &TreeNode{Val: 0, Left: stick(n/2, true), Right: stick(n/2, false)}
	within(t, func() { got = Diameter(fork) })
	if got != n {
		t.Errorf("Diameter(вилка из двух палок по %d) = %d, want %d", n/2, got, n)
	}
}
