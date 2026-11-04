package l04pointers

// Node это узел односвязного списка. Последний узел имеет Next == nil,
// пустой список это просто nil-указатель на голову.
type Node struct {
	Val  int
	Next *Node
}

// FromSlice строит список из значений в том же порядке. Пустой ввод: nil.
func FromSlice(vals []int) *Node {
	var head *Node
	for i := len(vals) - 1; i >= 0; i-- {
		PushFront(&head, vals[i])
	}
	return head
}

// ToSlice собирает значения списка по порядку. Пустой список: пустой слайс (длины 0).
func ToSlice(head *Node) []int {
	res := []int{}
	for n := head; n != nil; n = n.Next {
		res = append(res, n.Val)
	}
	return res
}

// PushFront добавляет значение в начало списка, на который указывает *head.
// Поэтому параметр это указатель на указатель: функция меняет переменную
// вызывающего, а не свою копию.
//
//	var list *Node
//	PushFront(&list, 1) // list: 1
//	PushFront(&list, 2) // list: 2 -> 1
func PushFront(head **Node, v int) {
	*head = &Node{Val: v, Next: *head}
}

// Reverse разворачивает список НА МЕСТЕ (без создания новых узлов)
// и возвращает новую голову. Reverse(nil) == nil.
func Reverse(head *Node) *Node {
	var prev *Node
	for cur := head; cur != nil; {
		next := cur.Next // запоминаем хвост, пока не перезаписали ссылку
		cur.Next = prev
		prev, cur = cur, next
	}
	return prev
}
