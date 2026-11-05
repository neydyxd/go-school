package l07linkedlists

// Middle возвращает средний узел списка. При чётной длине из двух
// средних берётся второй: у 1 -> 2 -> 3 -> 4 это узел 3. Для пустого
// списка nil. Список не меняй.
//
// Ожидаемая сложность: O(n) за один проход, O(1) памяти. Подсказка:
// быстрый и медленный указатели.
func Middle(head *ListNode) *ListNode {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
	}
	return slow
}
