package l04generics

import "cmp"

// MaxOf возвращает наибольший элемент и true. Для пустого ввода:
// нулевое значение T и false. Работает для любых упорядоченных типов:
// целых, вещественных, строк и типов на их основе.
func MaxOf[T cmp.Ordered](xs ...T) (T, bool) {
	panic("TODO: реализуй MaxOf")
}
