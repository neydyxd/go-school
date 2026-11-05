package l10graphs

// CanFinish сообщает, можно ли пройти все курсы 0..numCourses-1.
// Пара {a, b} в prerequisites значит «чтобы взять курс a, сначала пройди
// курс b», то есть ребро b -> a. Пройти всё можно, если в этом
// ориентированном графе нет цикла. Пара {a, a} это цикл из одного курса.
//
// Ожидаемая сложность: O(V + E). Помни, что «ромб» (два пути в одну
// вершину) это не цикл, и что DFS без запоминания уже проверенных вершин
// на графе с большим числом путей работает экспоненциально долго.
func CanFinish(numCourses int, prerequisites [][2]int) bool {
	adj := make([][]int, numCourses)
	for _, p := range prerequisites {
		adj[p[1]] = append(adj[p[1]], p[0])
	}
	// три цвета: 0 белая (не были), 1 серая (в текущем пути DFS),
	// 2 чёрная (полностью проверена, цикла через неё нет)
	state := make([]int8, numCourses)
	var hasCycle func(u int) bool
	hasCycle = func(u int) bool {
		state[u] = 1
		for _, v := range adj[u] {
			if state[v] == 1 {
				return true // ребро назад в текущий путь
			}
			if state[v] == 0 && hasCycle(v) {
				return true
			}
		}
		state[u] = 2
		return false
	}
	for u := range numCourses {
		if state[u] == 0 && hasCycle(u) {
			return false
		}
	}
	return true
}
