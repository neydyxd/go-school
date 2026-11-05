package l10graphs

// ShortestPathBinaryMatrix возвращает длину кратчайшего пути из левой
// верхней клетки в правую нижнюю в квадратной сетке n x n: 0 это свободная
// клетка, 1 стена. Ходить можно в любую из 8 соседних клеток (включая
// диагонали). Длина пути считается в клетках, включая начальную и
// конечную, поэтому для сетки 1x1 со свободной клеткой ответ 1. Если пути
// нет (в том числе когда старт или финиш это стена), верни -1.
//
// Ожидаемая сложность: O(n^2), обход в ширину. Поиск в глубину с
// перебором путей экспоненциален и не уложится в лимит.
func ShortestPathBinaryMatrix(grid [][]int) int {
	n := len(grid)
	if n == 0 || grid[0][0] == 1 || grid[n-1][n-1] == 1 {
		return -1
	}
	dist := make([][]int, n) // 0 значит «ещё не были»
	for i := range dist {
		dist[i] = make([]int, n)
	}
	dist[0][0] = 1
	queue := [][2]int{{0, 0}}
	for len(queue) > 0 {
		c := queue[0]
		queue = queue[1:]
		if c[0] == n-1 && c[1] == n-1 {
			return dist[c[0]][c[1]]
		}
		for dr := -1; dr <= 1; dr++ {
			for dc := -1; dc <= 1; dc++ {
				r, k := c[0]+dr, c[1]+dc
				if r < 0 || r >= n || k < 0 || k >= n || grid[r][k] == 1 || dist[r][k] != 0 {
					continue
				}
				dist[r][k] = dist[c[0]][c[1]] + 1
				queue = append(queue, [2]int{r, k})
			}
		}
	}
	return -1
}
