package l08sorting

import (
	"slices"
	"strconv"
	"strings"
)

// LargestNumber переставляет неотрицательные числа так, чтобы их запись
// подряд дала наибольшее число, и возвращает его строкой.
// [3 30 34 5 9] даёт "9534330", [0 0] даёт "0" (а не "00").
// Пустой вход даёт "". Склейка может быть длиннее любого int, поэтому
// строка.
//
// Подсказка: нужен свой порядок сравнения, а не сравнение чисел.
func LargestNumber(nums []int) string {
	strs := make([]string, len(nums))
	for i, v := range nums {
		strs[i] = strconv.Itoa(v)
	}
	// a раньше b, если склейка a+b больше b+a; такой порядок транзитивен
	slices.SortFunc(strs, func(a, b string) int { return strings.Compare(b+a, a+b) })
	if len(strs) > 0 && strs[0] == "0" {
		return "0" // первым стоит ноль, значит, все нули
	}
	return strings.Join(strs, "")
}
