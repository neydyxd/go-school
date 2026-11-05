package l10graphs

// NumIslands считает острова в сетке: '1' это суша, '0' вода. Остров это
// группа клеток суши, соединённых по горизонтали или вертикали (не по
// диагонали). Сетка может быть пустой; строки одной длины. Менять grid
// можно (например, «затапливать» посещённую сушу).
//
// Ожидаемая сложность: O(rows*cols). Сетка в большом тесте 1000x1000.
func NumIslands(grid [][]byte) int {
	count := 0
	for i := range grid {
		for j := range grid[i] {
			if grid[i][j] != '1' {
				continue
			}
			count++
			// затапливаем остров обходом в ширину: глубина рекурсии
			// на сплошной суше 1000x1000 доходила бы до миллиона
			grid[i][j] = '0'
			queue := [][2]int{{i, j}}
			for len(queue) > 0 {
				c := queue[0]
				queue = queue[1:]
				for _, d := range [4][2]int{{1, 0}, {-1, 0}, {0, 1}, {0, -1}} {
					r, k := c[0]+d[0], c[1]+d[1]
					if r >= 0 && r < len(grid) && k >= 0 && k < len(grid[r]) && grid[r][k] == '1' {
						grid[r][k] = '0'
						queue = append(queue, [2]int{r, k})
					}
				}
			}
		}
	}
	return count
}
