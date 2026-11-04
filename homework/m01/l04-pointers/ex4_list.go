package l04pointers

// Node это узел односвязного списка. Последний узел имеет Next == nil,
// пустой список это просто nil-указатель на голову.
type Node struct {
	Val  int
	Next *Node
}

// FromSlice строит список из значений в том же порядке. Пустой ввод: nil.
func FromSlice(vals []int) *Node {
	panic("TODO: реализуй FromSlice")
}

// ToSlice собирает значения списка по порядку. Пустой список: пустой слайс (длины 0).
func ToSlice(head *Node) []int {
	panic("TODO: реализуй ToSlice")
}

// PushFront добавляет значение в начало списка, на который указывает *head.
// Поэтому параметр это указатель на указатель: функция меняет переменную
// вызывающего, а не свою копию.
//
//	var list *Node
//	PushFront(&list, 1) // list: 1
//	PushFront(&list, 2) // list: 2 -> 1
func PushFront(head **Node, v int) {
	panic("TODO: реализуй PushFront")
}

// Reverse разворачивает список НА МЕСТЕ (без создания новых узлов)
// и возвращает новую голову. Reverse(nil) == nil.
func Reverse(head *Node) *Node {
	panic("TODO: реализуй Reverse")
}
