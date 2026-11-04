package l05slices

// NewMatrix создаёт матрицу rows x cols из нулей ровно за две аллокации:
// одна под общий массив всех rows*cols чисел, вторая под слайс строк.
// Строки нарезаются из общего массива.
//
// У каждой строки len == cap == cols, чтобы append к строке
// не залез в память следующей строки.
// rows <= 0 или cols <= 0: вернуть nil.
func NewMatrix(rows, cols int) [][]float64 {
	if rows <= 0 || cols <= 0 {
		return nil
	}
	cells := make([]float64, rows*cols)
	m := make([][]float64, rows)
	for i := range m {
		m[i] = cells[i*cols : (i+1)*cols : (i+1)*cols]
	}
	return m
}
