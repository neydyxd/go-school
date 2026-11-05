package l11heap

// MinHeap это двоичная куча минимумов на слайсе, написанная руками.
// Не используй container/heap, sort и slices.Sort: цель задания в том,
// чтобы самому написать просеивание вверх (sift up) и вниз (sift down).
// Дети узла i лежат в 2i+1 и 2i+2, родитель в (i-1)/2.
//
// Нулевое значение MinHeap это пустая куча, готовая к работе.
// Pop и Peek на пустой куче возвращают (0, false).
//
// Ожидаемая сложность: Push и Pop за O(log n), Peek и Len за O(1).
// Отсортированный слайс со вставкой на место (O(n) на Push) или поиск
// минимума перебором (O(n) на Pop) не уложатся в лимит на 1 000 000 операций.
type MinHeap struct {
	data []int
}

func (h *MinHeap) Len() int {
	panic("TODO: реализуй Len")
}

func (h *MinHeap) Push(x int) {
	panic("TODO: реализуй Push")
}

func (h *MinHeap) Pop() (int, bool) {
	panic("TODO: реализуй Pop")
}

func (h *MinHeap) Peek() (int, bool) {
	panic("TODO: реализуй Peek")
}
