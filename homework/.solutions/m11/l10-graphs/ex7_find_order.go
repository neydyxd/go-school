package l10graphs

// FindOrder возвращает порядок прохождения курсов 0..numCourses-1, в
// котором каждый курс идёт после всех своих пререквизитов (пара {a, b}
// значит «b раньше a»). Подходит любой верный порядок. Если порядка нет
// (есть цикл), верни пустой слайс или nil.
//
// Ожидаемая сложность: O(V + E), алгоритм Кана или DFS. Поиск вершины
// с нулевой степенью захода перебором на каждом шаге стоит O(V^2) и на
// 200 000 курсов не уложится в лимит.
func FindOrder(numCourses int, prerequisites [][2]int) []int {
	adj := make([][]int, numCourses)
	indeg := make([]int, numCourses)
	for _, p := range prerequisites {
		adj[p[1]] = append(adj[p[1]], p[0])
		indeg[p[0]]++
	}
	// алгоритм Кана: очередь вершин, у которых все пререквизиты пройдены
	order := make([]int, 0, numCourses)
	for u := range numCourses {
		if indeg[u] == 0 {
			order = append(order, u)
		}
	}
	// order одновременно ответ и очередь: голова очереди это индекс i
	for i := 0; i < len(order); i++ {
		for _, v := range adj[order[i]] {
			indeg[v]--
			if indeg[v] == 0 {
				order = append(order, v)
			}
		}
	}
	if len(order) < numCourses {
		return nil // оставшиеся вершины лежат на цикле или за ним
	}
	return order
}
