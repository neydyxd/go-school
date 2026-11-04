package l02controlflow

// Find ищет target в таблице grid построчно (сначала строка 0 слева направо,
// потом строка 1 и так далее) и возвращает координаты ПЕРВОГО вхождения.
// Если не нашлось (или таблица пустая), возвращает (-1, -1, false).
// Строки могут быть разной длины, в том числе пустыми.
// Попробуй выйти сразу из обоих циклов через метку и break.
func Find(grid [][]int, target int) (row, col int, ok bool) {
	row, col = -1, -1
search:
	for i, line := range grid {
		for j, v := range line {
			if v == target {
				row, col, ok = i, j, true
				break search // голый break вышел бы только из внутреннего цикла
			}
		}
	}
	return row, col, ok
}
