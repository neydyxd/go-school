package l11heap

import "container/heap"

// TopKFrequent возвращает k самых частых значений из nums в любом порядке.
// Гарантируется, что ответ единственный: k-е по частоте значение
// встречается строго чаще (k+1)-го. 1 <= k <= числа различных значений.
//
// Ожидаемая сложность: O(n log k) с кучей размера k (или O(n) корзинами
// по частоте). k раз искать самое частое среди оставшихся это O(k*d),
// где d число различных значений, и на большом тесте не уложится в лимит.
func TopKFrequent(nums []int, k int) []int {
	freq := make(map[int]int)
	for _, v := range nums {
		freq[v]++
	}
	// куча минимумов по частоте размера k: на вершине самый редкий из
	// кандидатов, его и вытесняем
	h := &freqHeap{}
	for v, c := range freq {
		heap.Push(h, pair{v, c})
		if h.Len() > k {
			heap.Pop(h)
		}
	}
	res := make([]int, 0, k)
	for _, p := range *h {
		res = append(res, p.val)
	}
	return res
}

type pair struct{ val, count int }

type freqHeap []pair

func (h freqHeap) Len() int           { return len(h) }
func (h freqHeap) Less(i, j int) bool { return h[i].count < h[j].count }
func (h freqHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *freqHeap) Push(x any)        { *h = append(*h, x.(pair)) }
func (h *freqHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}
