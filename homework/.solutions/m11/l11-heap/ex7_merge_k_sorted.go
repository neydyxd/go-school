package l11heap

import "container/heap"

// MergeKSorted сливает k отсортированных по возрастанию слайсов в один
// отсортированный. Слайсы могут быть пустыми или nil, lists тоже может
// быть пустым; тогда верни пустой слайс или nil. Входные слайсы не меняй.
//
// Цель задания: куча из k «голов» (текущих элементов каждого слайса) на
// container/heap, O(N log k), где N общее число элементов. Просмотр всех
// k голов на каждом шаге стоит O(N*k) и не уложится в лимит. Склеить всё
// и отсортировать нельзя: так не слить k потоков, которые не помещаются
// в память, поэтому тест запрещает импорт sort и slices в этом файле.
func MergeKSorted(lists [][]int) []int {
	total := 0
	h := &headHeap{lists: lists}
	for i, l := range lists {
		total += len(l)
		if len(l) > 0 {
			h.items = append(h.items, head{list: i, pos: 0})
		}
	}
	heap.Init(h)
	res := make([]int, 0, total)
	for h.Len() > 0 {
		top := h.items[0]
		res = append(res, lists[top.list][top.pos])
		if top.pos+1 < len(lists[top.list]) {
			h.items[0].pos++ // следующий элемент того же списка
			heap.Fix(h, 0)
		} else {
			heap.Pop(h) // список кончился
		}
	}
	return res
}

// head указывает на текущий элемент списка list.
type head struct{ list, pos int }

type headHeap struct {
	lists [][]int
	items []head
}

func (h *headHeap) val(i int) int      { return h.lists[h.items[i].list][h.items[i].pos] }
func (h *headHeap) Len() int           { return len(h.items) }
func (h *headHeap) Less(i, j int) bool { return h.val(i) < h.val(j) }
func (h *headHeap) Swap(i, j int)      { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *headHeap) Push(x any)         { h.items = append(h.items, x.(head)) }
func (h *headHeap) Pop() any {
	x := h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]
	return x
}
