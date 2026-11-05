package l07linkedlists

// CycleStart возвращает узел, с которого начинается цикл, или nil, если
// цикла нет. У 1 -> 2 -> 3 -> 4 -> (снова 2) это узел 2. Список не меняй.
//
// Ожидаемая сложность: O(n) по времени и O(1) памяти. Подсказка: после
// встречи быстрого и медленного указателей поставь один из них на голову
// и двигай оба по одному шагу; они встретятся в начале цикла.
func CycleStart(head *ListNode) *ListNode {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
		if slow == fast {
			// от головы до начала цикла столько же шагов, сколько от
			// точки встречи до начала цикла (по модулю длины цикла)
			for p := head; p != slow; {
				p, slow = p.Next, slow.Next
			}
			return slow
		}
	}
	return nil
}
