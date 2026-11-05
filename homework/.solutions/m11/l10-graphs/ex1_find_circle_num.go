package l10graphs

// FindCircleNum считает компоненты связности неориентированного графа,
// заданного матрицей смежности: isConnected[i][j] == 1, если города i и j
// соединены напрямую. Матрица симметрична, на диагонали единицы.
// Компонента («провинция») это группа городов, связанных напрямую или
// через другие города. Для n == 0 ответ 0.
//
// Ожидаемая сложность: O(n^2), то есть один просмотр матрицы. Транзитивное
// замыкание тройным циклом (O(n^3)) на n = 2000 не уложится в лимит.
func FindCircleNum(isConnected [][]int) int {
	n := len(isConnected)
	visited := make([]bool, n)
	count := 0
	for start := range n {
		if visited[start] {
			continue
		}
		count++
		visited[start] = true
		stack := []int{start}
		for len(stack) > 0 {
			u := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for v := range n { // строка матрицы: все потенциальные соседи
				if isConnected[u][v] == 1 && !visited[v] {
					visited[v] = true
					stack = append(stack, v)
				}
			}
		}
	}
	return count
}
