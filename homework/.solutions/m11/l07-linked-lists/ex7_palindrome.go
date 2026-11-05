package l07linkedlists

// IsPalindromeList сообщает, читается ли список одинаково в обе стороны:
// 1 -> 2 -> 2 -> 1 да, 1 -> 2 нет. Пустой список и список из одного узла
// палиндромы.
//
// Ожидаемая сложность: O(n) по времени и O(1) дополнительной памяти:
// найди середину, разверни вторую половину, сравни, разверни обратно.
// После вызова список должен быть ровно таким же, как до: те же узлы
// в том же порядке. Копирование значений в слайс не пройдёт тест
// на аллокации.
func IsPalindromeList(head *ListNode) bool {
	if head == nil || head.Next == nil {
		return true
	}
	// slow остановится на последнем узле первой половины
	slow, fast := head, head
	for fast.Next != nil && fast.Next.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
	}
	second := Reverse(slow.Next)
	ok := true
	for p, q := head, second; q != nil; p, q = p.Next, q.Next {
		if p.Val != q.Val {
			ok = false
			break
		}
	}
	slow.Next = Reverse(second) // возвращаем список в исходный вид
	return ok
}
