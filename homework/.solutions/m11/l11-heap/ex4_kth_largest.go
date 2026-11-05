package l11heap

import "container/heap"

// FindKthLargest возвращает k-й по величине элемент nums (в порядке
// убывания, с учётом повторов: в [3 3 1] второй по величине это 3).
// Гарантируется 1 <= k <= len(nums). Менять nums можно.
//
// Цель задания: куча минимумов размера k из container/heap, O(n log k)
// по времени и O(k) памяти. Сортировка всего слайса (O(n log n)) тоже
// пройдёт тесты, но на собесе спросят про кучу. k проходов «найди
// максимум и выкинь» стоят O(n*k) и не уложатся в лимит.
func FindKthLargest(nums []int, k int) int {
	// в куче держим k самых больших из просмотренных; её минимум и есть
	// k-й по величине
	h := make(intMinHeap, 0, k)
	for _, v := range nums {
		if h.Len() < k {
			heap.Push(&h, v)
		} else if v > h[0] {
			h[0] = v
			heap.Fix(&h, 0) // заменить корень дешевле, чем Pop + Push
		}
	}
	return h[0]
}

// intMinHeap реализует heap.Interface для кучи минимумов из int.
type intMinHeap []int

func (h intMinHeap) Len() int           { return len(h) }
func (h intMinHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h intMinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *intMinHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *intMinHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}
