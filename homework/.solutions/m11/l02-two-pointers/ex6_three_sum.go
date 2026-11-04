package l02twopointers

import "slices"

// ThreeSum возвращает все различные тройки чисел из nums с суммой 0.
// Каждая тройка упорядочена по возрастанию ([-1 -1 2], а не [2 -1 -1]),
// одинаковых троек в ответе быть не должно, порядок троек любой.
// Если троек нет, верни пустой слайс или nil. Переставлять nums можно.
//
// Ожидаемая сложность: O(n^2) по времени. Три вложенных цикла O(n^3)
// не уложатся в лимит на большом входе.
func ThreeSum(nums []int) [][]int {
	slices.Sort(nums)
	var res [][]int
	for i := 0; i < len(nums)-2; i++ {
		if i > 0 && nums[i] == nums[i-1] {
			continue // ту же первую цифру уже разобрали
		}
		l, r := i+1, len(nums)-1
		for l < r {
			sum := nums[i] + nums[l] + nums[r]
			switch {
			case sum < 0:
				l++
			case sum > 0:
				r--
			default:
				res = append(res, []int{nums[i], nums[l], nums[r]})
				l++
				for l < r && nums[l] == nums[l-1] {
					l++ // пропускаем повторы второго числа
				}
				r--
			}
		}
	}
	return res
}
