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
	return len(h.data)
}

func (h *MinHeap) Push(x int) {
	h.data = append(h.data, x)
	// sift up: поднимаем новый элемент, пока он меньше родителя
	i := len(h.data) - 1
	for i > 0 {
		p := (i - 1) / 2
		if h.data[p] <= h.data[i] {
			break
		}
		h.data[p], h.data[i] = h.data[i], h.data[p]
		i = p
	}
}

func (h *MinHeap) Pop() (int, bool) {
	if len(h.data) == 0 {
		return 0, false
	}
	top := h.data[0]
	last := len(h.data) - 1
	h.data[0] = h.data[last] // последний лист на место корня
	h.data = h.data[:last]
	siftDown(h.data, 0, len(h.data))
	return top, true
}

func (h *MinHeap) Peek() (int, bool) {
	if len(h.data) == 0 {
		return 0, false
	}
	return h.data[0], true
}

// siftDown опускает a[i] в куче минимумов a[:n], пока он больше
// меньшего из детей.
func siftDown(a []int, i, n int) {
	for {
		small := i
		if l := 2*i + 1; l < n && a[l] < a[small] {
			small = l
		}
		if r := 2*i + 2; r < n && a[r] < a[small] {
			small = r
		}
		if small == i {
			return
		}
		a[i], a[small] = a[small], a[i]
		i = small
	}
}
