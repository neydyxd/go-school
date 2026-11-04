package l04generics

// Map применяет f к каждому элементу s и возвращает новый слайс результатов
// той же длины и в том же порядке. Входной слайс не меняется.
func Map[T, U any](s []T, f func(T) U) []U {
	panic("TODO: реализуй Map")
}

// Filter возвращает НОВЫЙ слайс из элементов s, для которых keep вернула true,
// в исходном порядке. Входной слайс (и его базовый массив) не меняется.
func Filter[T any](s []T, keep func(T) bool) []T {
	panic("TODO: реализуй Filter")
}

// Reduce сворачивает s слева направо: acc = f(acc, s[0]), затем f(acc, s[1])...
// Начальное значение acc: init. Пустой слайс: init.
func Reduce[T, A any](s []T, init A, f func(A, T) A) A {
	panic("TODO: реализуй Reduce")
}
