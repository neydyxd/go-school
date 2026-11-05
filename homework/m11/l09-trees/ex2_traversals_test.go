package l09trees

import (
	"slices"
	"testing"
)

func TestTraversals(t *testing.T) {
	tests := []struct {
		name           string
		in             string
		pre, ino, post []int
	}{
		{"пустое дерево", "", nil, nil, nil},
		{"один узел", "1", []int{1}, []int{1}, []int{1}},
		{"LeetCode", "1 # 2 3", []int{1, 2, 3}, []int{1, 3, 2}, []int{3, 2, 1}},
		{"полное дерево", "1 2 3 4 5 6 7",
			[]int{1, 2, 4, 5, 3, 6, 7}, []int{4, 2, 5, 1, 6, 3, 7}, []int{4, 5, 2, 6, 7, 3, 1}},
		{"палка влево", "3 2 # 1", []int{3, 2, 1}, []int{1, 2, 3}, []int{1, 2, 3}},
		{"палка вправо", "1 # 2 # 3", []int{1, 2, 3}, []int{1, 2, 3}, []int{3, 2, 1}},
		{"неполное", "1 2 3 # 4 5 # # # 6", []int{1, 2, 4, 3, 5, 6}, []int{2, 4, 1, 6, 5, 3}, []int{4, 2, 6, 5, 3, 1}},
	}
	type fn struct {
		name string
		f    func(*TreeNode) []int
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, f := range []fn{{"Preorder", Preorder}, {"Inorder", Inorder}, {"Postorder", Postorder}} {
				want := [][]int{tt.pre, tt.ino, tt.post}[i]
				got := f.f(tree(tt.in))
				if len(got) != len(want) || (len(got) > 0 && !slices.Equal(got, want)) {
					t.Errorf("%s(%q) = %v, want %v", f.name, tt.in, got, want)
				}
			}
		})
	}
}

func TestTraversalsBig(t *testing.T) {
	const n = 100_000
	root := stick(n, true) // n, n-1, ..., 1 влево
	var pre, ino, post []int
	within(t, func() {
		pre, ino, post = Preorder(root), Inorder(root), Postorder(root)
	})
	if len(pre) != n || pre[0] != n || pre[n-1] != 1 {
		t.Errorf("Preorder(палка влево из %d): len=%d, неверный порядок", n, len(pre))
	}
	if len(ino) != n || ino[0] != 1 || ino[n-1] != n {
		t.Errorf("Inorder(палка влево из %d): len=%d, неверный порядок", n, len(ino))
	}
	if len(post) != n || post[0] != 1 || post[n-1] != n {
		t.Errorf("Postorder(палка влево из %d): len=%d, неверный порядок", n, len(post))
	}
}
