package l07linkedlists

// MergeTwo сливает два списка, отсортированных по неубыванию, в один
// отсортированный и возвращает его голову. Используй существующие узлы,
// новых не создавай. При равных значениях первым идёт узел из a.
// Любой из списков (или оба) может быть пустым.
//
// Ожидаемая сложность: O(n + m) по времени, O(1) дополнительной памяти.
func MergeTwo(a, b *ListNode) *ListNode {
	var dummy ListNode // фиктивная голова: не нужна отдельная ветка для первого узла
	tail := &dummy
	for a != nil && b != nil {
		if a.Val <= b.Val {
			tail.Next, a = a, a.Next
		} else {
			tail.Next, b = b, b.Next
		}
		tail = tail.Next
	}
	if a != nil { // один из списков кончился, остаток второго цепляем целиком
		tail.Next = a
	} else {
		tail.Next = b
	}
	return dummy.Next
}
