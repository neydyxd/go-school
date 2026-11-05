package l09trees

// Search ищет в BST узел со значением val и возвращает указатель на него
// или nil, если такого значения нет.
//
// Insert добавляет val в BST и возвращает корень (для пустого дерева
// это новый узел). Если val уже есть, дерево не меняется. Существующие
// узлы не пересоздавай: новый узел становится листом.
//
// Ожидаемая сложность: O(h), где h это высота дерева. Обойти всё дерево
// в поисках значения (O(n)) не уложится в лимит на 200 000 запросов.
func Search(root *TreeNode, val int) *TreeNode {
	for root != nil && root.Val != val {
		if val < root.Val {
			root = root.Left
		} else {
			root = root.Right
		}
	}
	return root
}

func Insert(root *TreeNode, val int) *TreeNode {
	nd := &TreeNode{Val: val}
	if root == nil {
		return nd
	}
	cur := root
	for {
		switch {
		case val < cur.Val:
			if cur.Left == nil {
				cur.Left = nd
				return root
			}
			cur = cur.Left
		case val > cur.Val:
			if cur.Right == nil {
				cur.Right = nd
				return root
			}
			cur = cur.Right
		default:
			return root // уже есть
		}
	}
}
