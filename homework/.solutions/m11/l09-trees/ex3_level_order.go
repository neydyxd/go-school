package l09trees

// LevelOrder возвращает значения узлов по уровням: сначала корень,
// потом его дети слева направо, потом внуки и так далее. Каждый уровень
// отдельным слайсом. Для пустого дерева верни пустой слайс или nil.
//
// Ожидаемая сложность: O(n) по времени. Подход «для каждой глубины d
// заново обойти дерево и собрать узлы глубины d» стоит O(n*h) и на
// «палке» из 100 000 узлов не уложится в лимит.
func LevelOrder(root *TreeNode) [][]int {
	if root == nil {
		return nil
	}
	var res [][]int
	level := []*TreeNode{root}
	for len(level) > 0 {
		vals := make([]int, 0, len(level))
		var next []*TreeNode
		for _, n := range level {
			vals = append(vals, n.Val)
			if n.Left != nil {
				next = append(next, n.Left)
			}
			if n.Right != nil {
				next = append(next, n.Right)
			}
		}
		res = append(res, vals)
		level = next
	}
	return res
}
