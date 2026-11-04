package l05slices

// NewMatrix создаёт матрицу rows x cols из нулей ровно за две аллокации:
// одна под общий массив всех rows*cols чисел, вторая под слайс строк.
// Строки нарезаются из общего массива.
//
// У каждой строки len == cap == cols, чтобы append к строке
// не залез в память следующей строки.
// rows <= 0 или cols <= 0: вернуть nil.
func NewMatrix(rows, cols int) [][]float64 {
	panic("TODO: реализуй NewMatrix")
}
