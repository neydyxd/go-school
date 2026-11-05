package l09trees

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestSearch(t *testing.T) {
	root := tree("8 4 12 2 6 10 14")
	for _, v := range []int{8, 4, 12, 2, 6, 10, 14} {
		if got := Search(root, v); got != find(root, v) {
			t.Errorf("Search(%d) вернула не тот узел: %v", v, got)
		}
	}
	for _, v := range []int{0, 5, 9, 15, -1} {
		if got := Search(root, v); got != nil {
			t.Errorf("Search(%d) = узел %d, want nil", v, got.Val)
		}
	}
	if got := Search(nil, 1); got != nil {
		t.Errorf("Search(nil, 1) = %v, want nil", got)
	}
}

func TestInsert(t *testing.T) {
	t.Run("в пустое дерево", func(t *testing.T) {
		root := Insert(nil, 5)
		if root == nil || root.Val != 5 || root.Left != nil || root.Right != nil {
			t.Fatalf("Insert(nil, 5) = %+v, want один узел 5", root)
		}
	})
	t.Run("последовательность", func(t *testing.T) {
		var root *TreeNode
		vals := []int{5, 3, 8, 1, 4, 7, 9, 2, 6}
		for _, v := range vals {
			root = Insert(root, v)
		}
		got := inorder(root)
		want := slices.Sorted(slices.Values(vals))
		if !slices.Equal(got, want) {
			t.Errorf("после Insert %v обход inorder = %v, want %v", vals, got, want)
		}
		if root.Val != 5 || root.Left.Val != 3 || root.Right.Val != 8 {
			t.Errorf("Insert перестроил дерево: корень %d, дети %d и %d, want 5, 3, 8", root.Val, root.Left.Val, root.Right.Val)
		}
	})
	t.Run("повтор не добавляется", func(t *testing.T) {
		root := tree("5 3 8")
		root = Insert(root, 3)
		root = Insert(root, 5)
		if n := size(root); n != 3 {
			t.Errorf("Insert существующих значений: узлов %d, want 3", n)
		}
	})
	t.Run("корень и узлы не пересоздаются", func(t *testing.T) {
		root := tree("5 3 8")
		left, right := root.Left, root.Right
		got := Insert(root, 4)
		if got != root || root.Left != left || root.Right != right {
			t.Fatalf("Insert(4) пересоздал узлы, want прежние указатели")
		}
		if left.Right == nil || left.Right.Val != 4 {
			t.Errorf("Insert(4): want 4 правым ребёнком узла 3")
		}
	})
}

func TestBSTBig(t *testing.T) {
	const n = 1<<18 - 1
	root := balanced(1, n)
	r := rand.New(rand.NewPCG(1, 2))
	queries := make([]int, 200_000)
	for i := range queries {
		queries[i] = r.IntN(n+10) - 5 // часть запросов мимо дерева
	}
	misses := 0
	within(t, func() {
		for _, q := range queries {
			nd := Search(root, q)
			if (nd == nil) != (q < 1 || q > n) || (nd != nil && nd.Val != q) {
				misses++
			}
		}
	})
	if misses > 0 {
		t.Errorf("Search на дереве из %d узлов: %d неверных ответов", n, misses)
	}

	var bst *TreeNode
	vals := r.Perm(200_000)
	within(t, func() {
		for _, v := range vals {
			bst = Insert(bst, v)
		}
	})
	if got := inorder(bst); len(got) != len(vals) || !slices.IsSorted(got) {
		t.Errorf("Insert 200 000 случайных значений: %d узлов, отсортировано %v", len(got), slices.IsSorted(got))
	}
}
