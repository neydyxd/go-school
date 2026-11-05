package l05stackqueue

// LargestRectangle возвращает площадь самого большого прямоугольника,
// который помещается в гистограмму из столбиков ширины 1 и высоты
// heights[i] (высоты неотрицательные). [2 1 5 6 2 3] даёт 10:
// прямоугольник высоты 5 на столбиках 5 и 6.
// Пустая гистограмма даёт 0.
//
// Ожидаемая сложность: O(n) через монотонный стек. Расширение влево
// и вправо от каждого столбика это O(n^2) на возрастающих высотах.
func LargestRectangle(heights []int) int {
	best := 0
	var st []int // индексы столбиков с неубывающими высотами
	for i := 0; i <= len(heights); i++ {
		h := 0 // фиктивный столбик высоты 0 в конце выталкивает всё
		if i < len(heights) {
			h = heights[i]
		}
		for len(st) > 0 && heights[st[len(st)-1]] >= h {
			top := st[len(st)-1]
			st = st[:len(st)-1]
			left := -1 // ближайший слева более низкий столбик
			if len(st) > 0 {
				left = st[len(st)-1]
			}
			best = max(best, heights[top]*(i-left-1))
		}
		st = append(st, i)
	}
	return best
}
