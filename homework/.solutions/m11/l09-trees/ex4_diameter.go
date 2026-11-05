package l09trees

// Diameter возвращает диаметр дерева: число рёбер на самом длинном пути
// между любыми двумя узлами. Путь не обязан проходить через корень.
// У пустого дерева и у одного узла диаметр 0.
//
// Ожидаемая сложность: O(n) по времени, один обход. Если для каждого
// узла заново считать высоту поддеревьев, выйдет O(n^2), и на «палке»
// из 100 000 узлов это не уложится в лимит.
func Diameter(root *TreeNode) int {
	best := 0
	// height возвращает высоту в узлах и попутно обновляет best:
	// самый длинный путь через n это высота слева плюс высота справа.
	var height func(n *TreeNode) int
	height = func(n *TreeNode) int {
		if n == nil {
			return 0
		}
		l, r := height(n.Left), height(n.Right)
		best = max(best, l+r)
		return 1 + max(l, r)
	}
	height(root)
	return best
}
