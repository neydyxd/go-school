package l11heap

import "container/heap"

// LastStoneWeight: камни с весами stones. Пока камней больше одного,
// берём два самых тяжёлых x <= y и сталкиваем: если x == y, исчезают оба,
// иначе остаётся камень весом y-x. Верни вес последнего камня или 0, если
// камней не осталось (в том числе для пустого входа). Менять stones можно.
//
// Ожидаемая сложность: O(n log n), куча максимумов на container/heap.
// Сортировка на каждом шаге или поиск двух максимумов перебором дают
// O(n^2) и больше и не уложатся в лимит на 200 000 камней.
func LastStoneWeight(stones []int) int {
	h := maxHeap(stones)
	heap.Init(&h) // O(n), просеивание снизу вверх
	for h.Len() > 1 {
		y := heap.Pop(&h).(int)
		x := heap.Pop(&h).(int)
		if y != x {
			heap.Push(&h, y-x)
		}
	}
	if h.Len() == 0 {
		return 0
	}
	return h[0]
}

// maxHeap: тот же слайс, но Less «наоборот» превращает кучу в кучу максимумов.
type maxHeap []int

func (h maxHeap) Len() int           { return len(h) }
func (h maxHeap) Less(i, j int) bool { return h[i] > h[j] }
func (h maxHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *maxHeap) Push(x any)        { *h = append(*h, x.(int)) }
func (h *maxHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}
