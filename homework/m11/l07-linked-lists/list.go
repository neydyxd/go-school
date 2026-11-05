package l07linkedlists

// ListNode это узел односвязного списка. Список задаётся указателем на
// первый узел (голову), пустой список это nil.
type ListNode struct {
	Val  int
	Next *ListNode
}
