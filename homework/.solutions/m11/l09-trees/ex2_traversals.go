package l09trees

// Три обхода в глубину. Каждая функция возвращает значения узлов
// в своём порядке; для пустого дерева верни пустой слайс или nil.
//
//	Preorder:  корень, левое поддерево, правое поддерево
//	Inorder:   левое поддерево, корень, правое поддерево
//	Postorder: левое поддерево, правое поддерево, корень
//
// Цель задания: написать все три без рекурсии, на явном стеке
// (слайс как стек): именно это просят на собесе. Тест разбирает файл
// через go/parser и падает, если функция или замыкание вызывает само себя.
//
// Ожидаемая сложность: O(n) по времени и O(h) дополнительной памяти,
// где h это высота дерева.
func Preorder(root *TreeNode) []int {
	var res []int
	stack := []*TreeNode{root}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if n == nil {
			continue
		}
		res = append(res, n.Val)
		// правого кладём первым, чтобы левый оказался на вершине
		stack = append(stack, n.Right, n.Left)
	}
	return res
}

func Inorder(root *TreeNode) []int {
	var res []int
	var stack []*TreeNode
	cur := root
	for cur != nil || len(stack) > 0 {
		for cur != nil { // спускаемся влево до упора, запоминая путь
			stack = append(stack, cur)
			cur = cur.Left
		}
		cur = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		res = append(res, cur.Val)
		cur = cur.Right
	}
	return res
}

func Postorder(root *TreeNode) []int {
	var res []int
	var stack []*TreeNode
	var prev *TreeNode // последний выведенный узел
	cur := root
	for cur != nil || len(stack) > 0 {
		for cur != nil {
			stack = append(stack, cur)
			cur = cur.Left
		}
		top := stack[len(stack)-1]
		if top.Right != nil && top.Right != prev {
			cur = top.Right // правое поддерево ещё не обойдено
			continue
		}
		res = append(res, top.Val)
		prev = top
		stack = stack[:len(stack)-1]
	}
	return res
}
