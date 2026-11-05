package l09trees

import (
	"math/rand/v2"
	"testing"
)

func TestLowestCommonAncestorBST(t *testing.T) {
	const in = "6 2 8 0 4 7 9 # # 3 5"
	tests := []struct {
		name       string
		p, q, want int
	}{
		{"по разные стороны корня", 2, 8, 6},
		{"предок сам себе", 2, 4, 2},
		{"q выше p", 5, 2, 2},
		{"глубоко в левом поддереве", 3, 5, 4},
		{"лист и далёкий лист", 0, 5, 2},
		{"p равен q", 7, 7, 7},
		{"корень и лист", 6, 9, 6},
		{"листья справа", 7, 9, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := tree(in)
			p, q := find(root, tt.p), find(root, tt.q)
			got := LowestCommonAncestorBST(root, p, q)
			if got != find(root, tt.want) {
				t.Errorf("LowestCommonAncestorBST(%q, %d, %d) = %v, want узел %d", in, tt.p, tt.q, val(got), tt.want)
			}
		})
	}
	t.Run("один узел", func(t *testing.T) {
		root := tree("1")
		if got := LowestCommonAncestorBST(root, root, root); got != root {
			t.Errorf("LowestCommonAncestorBST(один узел) = %v, want корень", val(got))
		}
	})
	t.Run("палка", func(t *testing.T) {
		root := stick(1000, false)
		p, q := find(root, 500), find(root, 1000)
		if got := LowestCommonAncestorBST(root, p, q); got != p {
			t.Errorf("LowestCommonAncestorBST(палка, 500, 1000) = %v, want 500", val(got))
		}
	})
}

// val печатает значение узла или nil для сообщений об ошибках.
func val(n *TreeNode) any {
	if n == nil {
		return nil
	}
	return n.Val
}

func TestLowestCommonAncestorBSTBig(t *testing.T) {
	const n = 1<<18 - 1
	root := balanced(1, n)
	r := rand.New(rand.NewPCG(3, 4))
	type pair struct{ p, q *TreeNode }
	pairs := make([]pair, 200_000)
	for i := range pairs {
		// узлы глубоко внизу, рядом друг с другом: ответ тоже внизу
		a := r.IntN(n-20) + 1
		pairs[i] = pair{bstFind(root, a), bstFind(root, a+r.IntN(20))}
	}
	bad := 0
	within(t, func() {
		for _, pr := range pairs {
			got := LowestCommonAncestorBST(root, pr.p, pr.q)
			if got == nil || got.Val < min(pr.p.Val, pr.q.Val) || got.Val > max(pr.p.Val, pr.q.Val) {
				bad++
			}
		}
	})
	if bad > 0 {
		t.Errorf("LowestCommonAncestorBST на дереве из %d узлов: %d неверных ответов", n, bad)
	}
}

// bstFind ищет узел в BST (тестовый помощник, не зависит от твоего Search).
func bstFind(root *TreeNode, v int) *TreeNode {
	for root != nil && root.Val != v {
		if v < root.Val {
			root = root.Left
		} else {
			root = root.Right
		}
	}
	return root
}
