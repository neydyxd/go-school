package l09trees

import "testing"

func TestIsValidBST(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"пустое дерево", "", true},
		{"один узел", "7", true},
		{"три узла", "2 1 3", true},
		{"правый ребёнок меньше", "5 1 4 # # 3 6", false},
		// 3 меньше корня 5, но лежит в его правом поддереве
		{"нарушение через уровень", "5 4 6 # # 3 7", false},
		{"нарушение глубоко слева", "10 5 15 1 12", false},
		{"повтор слева", "2 2", false},
		{"повтор справа", "2 # 2", false},
		{"палка вправо", "1 # 2 # 3", true},
		{"MinInt в корне", "-9223372036854775808", true},
		{"MaxInt в корне", "9223372036854775807", true},
		{"MinInt и MaxInt", "0 -9223372036854775808 9223372036854775807", true},
		{"MaxInt и повтор MaxInt", "9223372036854775807 # 9223372036854775807", false},
		{"большое корректное", "8 4 12 2 6 10 14 1 3 5 7 9 11 13 15", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsValidBST(tree(tt.in)); got != tt.want {
				t.Errorf("IsValidBST(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsValidBSTBig(t *testing.T) {
	const n = 100_000
	for _, left := range []bool{false, true} {
		root := stick(n, left)
		var got bool
		within(t, func() { got = IsValidBST(root) })
		if !got {
			t.Errorf("IsValidBST(палка-BST из %d, влево=%v) = false, want true", n, left)
		}
	}
	// портим самый глубокий узел палки: значение 0 справа от 1, 2, ...
	root := stick(n, false)
	cur := root
	for cur.Right != nil {
		cur = cur.Right
	}
	cur.Right = &TreeNode{Val: 0}
	var got bool
	within(t, func() { got = IsValidBST(root) })
	if got {
		t.Errorf("IsValidBST(палка с 0 в самом низу) = true, want false")
	}
}
