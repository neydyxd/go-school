package l10graphs

import "container/heap"

// NetworkDelayTime: сеть из n узлов с номерами 1..n, times[i] = {u, v, w}
// это направленное ребро u -> v, сигнал идёт по нему w миллисекунд
// (0 <= w <= 100). Из узла k отправили сигнал. Верни время, за которое
// его получат все узлы, то есть максимум из кратчайших расстояний от k.
// Если какой-то узел недостижим, верни -1.
//
// Ожидаемая сложность: O((V + E) log V), алгоритм Дейкстры с кучей
// из container/heap. Дейкстра с поиском минимума перебором (O(V^2)) или
// Беллман-Форд (O(V*E)) на 200 000 узлов не уложатся в лимит.
func NetworkDelayTime(times [][3]int, n, k int) int {
	type edge struct{ to, w int }
	adj := make([][]edge, n+1)
	for _, t := range times {
		adj[t[0]] = append(adj[t[0]], edge{t[1], t[2]})
	}
	const inf = int(^uint(0) >> 1)
	dist := make([]int, n+1)
	for i := range dist {
		dist[i] = inf
	}
	dist[k] = 0
	pq := &distHeap{{node: k, d: 0}}
	for pq.Len() > 0 {
		cur := heap.Pop(pq).(item)
		if cur.d > dist[cur.node] {
			continue // устаревшая запись: узел уже достали с меньшим d
		}
		for _, e := range adj[cur.node] {
			if nd := cur.d + e.w; nd < dist[e.to] {
				dist[e.to] = nd
				heap.Push(pq, item{node: e.to, d: nd})
			}
		}
	}
	ans := 0
	for _, d := range dist[1:] {
		if d == inf {
			return -1
		}
		ans = max(ans, d)
	}
	return ans
}

type item struct{ node, d int }

// distHeap это min-куча по расстоянию для container/heap.
type distHeap []item

func (h distHeap) Len() int           { return len(h) }
func (h distHeap) Less(i, j int) bool { return h[i].d < h[j].d }
func (h distHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *distHeap) Push(x any)        { *h = append(*h, x.(item)) }
func (h *distHeap) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}
