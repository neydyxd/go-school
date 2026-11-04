package l05slices

import "slices"

// RotateLeft сдвигает элементы s по кругу влево на k позиций, на месте.
// [1 2 3 4 5], k=2 -> [3 4 5 1 2].
//
// k может быть больше len(s) и может быть отрицательным
// (отрицательный k означает сдвиг вправо).
// Дополнительную память не выделяй: только перестановки внутри s.
// Подсказка: три разворота.
// Пустой и nil слайс: ничего не делать.
func RotateLeft(s []int, k int) {
	n := len(s)
	if n == 0 {
		return
	}
	k = ((k % n) + n) % n
	// [a b | c d e] -> [b a | e d c] -> [c d e a b]
	slices.Reverse(s[:k])
	slices.Reverse(s[k:])
	slices.Reverse(s)
}
