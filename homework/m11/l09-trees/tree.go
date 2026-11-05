package l09trees

// TreeNode это узел двоичного дерева, как в задачах LeetCode.
// Пустое дерево это nil.
type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}
