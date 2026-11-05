package l09trees

import "testing"

func TestLowestCommonAncestor(t *testing.T) {
	const in = "3 5 1 6 2 0 8 # # 7 4"
	tests := []struct {
		name       string
		p, q, want int
	}{
		{"по разные стороны корня", 5, 1, 3},
		{"предок сам себе", 5, 4, 5},
		{"оба глубоко", 7, 4, 2},
		{"лист и лист в разных ветках", 6, 4, 5},
		{"p равен q", 8, 8, 8},
		{"корень", 3, 7, 3},
		{"q выше p", 4, 5, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tree(in)
			p, q := find(root, tt.p), find(root, tt.q)
			got := LowestCommonAncestor(root, p, q)
			if got != find(root, tt.want) {
				t.Errorf("LowestCommonAncestor(%q, %d, %d) = %v, want узел %d", in, tt.p, tt.q, val(got), tt.want)
			}
		})
	}
	t.Run("одинаковые значения", func(t *testing.T) {
		// все узлы со значением 1: сравнение по значению даст корень
		root := tree("1 1 1 1 1")
		p, q := root.Left.Left, root.Left.Right
		if got := LowestCommonAncestor(root, p, q); got != root.Left {
			t.Errorf("LowestCommonAncestor(все значения 1, внуки слева) вернул не левого ребёнка корня")
		}
	})
	t.Run("один узел", func(t *testing.T) {
		root := tree("1")
		if got := LowestCommonAncestor(root, root, root); got != root {
			t.Errorf("LowestCommonAncestor(один узел) = %v, want корень", val(got))
		}
	})
}

func TestLowestCommonAncestorBig(t *testing.T) {
	const n = 100_000
	root := stick(n, true) // n, n-1, ..., 1 влево
	p, q := find(root, 2), find(root, 1)
	var got *TreeNode
	within(t, func() { got = LowestCommonAncestor(root, p, q) })
	if got != p {
		t.Errorf("LowestCommonAncestor(палка из %d, 2, 1) = %v, want 2", n, val(got))
	}
	// вилка: p и q на концах двух длинных палок
	fork := &TreeNode{Val: 0, Left: stick(n/2, true), Right: stick(n/2, false)}
	p, q = find(fork.Left, 1), find(fork.Right, n/2)
	within(t, func() { got = LowestCommonAncestor(fork, p, q) })
	if got != fork {
		t.Errorf("LowestCommonAncestor(вилка, концы палок) = %v, want корень 0", val(got))
	}
}
