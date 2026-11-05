package l09trees

// MaxDepth возвращает высоту дерева: число узлов на самом длинном пути
// от корня до листа. У пустого дерева (nil) высота 0, у одного узла 1.
//
// Ожидаемая сложность: O(n) по времени. Дерево может быть «палкой»
// глубиной 100 000 узлов.
func MaxDepth(root *TreeNode) int {
	if root == nil {
		return 0
	}
	return 1 + max(MaxDepth(root.Left), MaxDepth(root.Right))
}
