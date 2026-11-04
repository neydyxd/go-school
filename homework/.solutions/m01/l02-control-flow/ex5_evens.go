package l02controlflow

import "iter"

// Evens возвращает итератор по чётным неотрицательным числам меньше limit:
// 0, 2, 4, ... Для limit <= 0 итератор пустой.
//
//	for v := range Evens(7) { ... } // 0 2 4 6
//
// Итератор обязан уважать break: если yield вернул false, сразу прекращай
// перебор. Иначе рантайм упадёт с паникой
// "range function continued iteration after function for loop body returned false".
func Evens(limit int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for v := 0; v < limit; v += 2 {
			if !yield(v) {
				return
			}
		}
	}
}
