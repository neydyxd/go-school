package l05slices

// Head возвращает первые n байт data как независимую копию.
//
// Результат не должен делить память с data: запись в результат не меняет data,
// а большой массив под data может быть собран сборщиком мусора, даже если
// результат живёт долго (например, лежит в кеше).
// n > len(data): копия всего data. n <= 0: слайс длины 0.
func Head(data []byte, n int) []byte {
	n = max(0, min(n, len(data)))
	out := make([]byte, n)
	copy(out, data)
	return out
}
