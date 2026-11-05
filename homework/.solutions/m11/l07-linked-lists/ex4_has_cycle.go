package l07linkedlists

// HasCycle сообщает, есть ли в списке цикл: Next какого-то узла
// указывает на узел, который уже встречался раньше. nil это список
// без цикла. Список не меняй.
//
// Ожидаемая сложность: O(n) по времени и O(1) памяти (алгоритм Флойда,
// «черепаха и заяц»). Решение с map посещённых узлов работает, но
// тратит O(n) памяти и не пройдёт тест на аллокации.
func HasCycle(head *ListNode) bool {
	slow, fast := head, head
	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
		if slow == fast {
			return true
		}
	}
	return false
}
