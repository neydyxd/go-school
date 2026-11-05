package l07linkedlists

// RemoveNthFromEnd удаляет n-й с конца узел и возвращает голову
// получившегося списка. n = 1 это последний узел, n = длина списка это
// голова. Гарантируется 1 <= n <= длина, список не пустой. Если в списке
// был один узел, верни nil.
//
// Ожидаемая сложность: O(n) за один проход, O(1) памяти. Подсказка:
// второй указатель отстаёт от первого ровно на n узлов.
func RemoveNthFromEnd(head *ListNode, n int) *ListNode {
	dummy := &ListNode{Next: head} // фиктивный узел перед головой: удаление головы не особый случай
	fast, slow := dummy, dummy
	for range n {
		fast = fast.Next
	}
	for fast.Next != nil {
		fast, slow = fast.Next, slow.Next
	}
	slow.Next = slow.Next.Next // slow стоит прямо перед удаляемым
	return dummy.Next
}
