package l10graphs

// ValidPath сообщает, есть ли путь из source в destination в
// неориентированном графе с вершинами 0..n-1 и рёбрами edges.
// В графе могут быть циклы и несколько компонент связности;
// source == destination означает, что путь есть (пустой).
//
// Ожидаемая сложность: O(V + E). Построй список смежности один раз.
// Обход без отметки посещённых зациклится, а просмотр всех рёбер на каждом
// шаге обхода (O(V*E)) не уложится в лимит на 200 000 вершин.
func ValidPath(n int, edges [][2]int, source, destination int) bool {
	adj := make([][]int, n)
	for _, e := range edges {
		adj[e[0]] = append(adj[e[0]], e[1])
		adj[e[1]] = append(adj[e[1]], e[0])
	}
	visited := make([]bool, n)
	visited[source] = true
	queue := []int{source}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		if u == destination {
			return true
		}
		for _, v := range adj[u] {
			if !visited[v] { // отмечаем при постановке в очередь
				visited[v] = true
				queue = append(queue, v)
			}
		}
	}
	return false
}
