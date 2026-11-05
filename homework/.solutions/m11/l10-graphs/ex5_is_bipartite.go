package l10graphs

// IsBipartite сообщает, можно ли раскрасить вершины неориентированного
// графа в два цвета так, чтобы концы каждого ребра были разного цвета.
// graph[u] это список соседей вершины u (вершины 0..len(graph)-1), петель
// и кратных рёбер нет. Граф может быть несвязным: тогда двудольной должна
// быть каждая компонента. Пустой граф двудольный.
//
// Ожидаемая сложность: O(V + E).
func IsBipartite(graph [][]int) bool {
	color := make([]int, len(graph)) // 0 не покрашена, 1 и -1 доли
	for start := range graph {       // каждую компоненту отдельно
		if color[start] != 0 {
			continue
		}
		color[start] = 1
		queue := []int{start}
		for len(queue) > 0 {
			u := queue[0]
			queue = queue[1:]
			for _, v := range graph[u] {
				switch color[v] {
				case 0:
					color[v] = -color[u]
					queue = append(queue, v)
				case color[u]:
					return false // ребро внутри одной доли
				}
			}
		}
	}
	return true
}
