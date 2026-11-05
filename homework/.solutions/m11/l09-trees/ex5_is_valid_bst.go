package l09trees

// IsValidBST сообщает, является ли дерево двоичным деревом поиска:
// для каждого узла ВСЕ значения левого поддерева строго меньше его
// значения, а ВСЕ значения правого строго больше. Повторы запрещены.
// Пустое дерево считается корректным BST. Значения узлов могут быть
// любыми int, включая math.MinInt и math.MaxInt.
//
// Ожидаемая сложность: O(n) по времени. Если для каждого узла искать
// максимум левого и минимум правого поддерева, выйдет O(n^2), и на
// «палке» из 100 000 узлов это не уложится в лимит.
func IsValidBST(root *TreeNode) bool {
	// Симметричный обход BST идёт строго по возрастанию. Сравниваем
	// с предыдущим узлом, а не с «минус бесконечностью»: так не нужен
	// сторожевой math.MinInt, который сломался бы на узле со значением MinInt.
	var stack []*TreeNode
	var prev *TreeNode
	cur := root
	for cur != nil || len(stack) > 0 {
		for cur != nil {
			stack = append(stack, cur)
			cur = cur.Left
		}
		cur = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if prev != nil && prev.Val >= cur.Val {
			return false
		}
		prev = cur
		cur = cur.Right
	}
	return true
}
