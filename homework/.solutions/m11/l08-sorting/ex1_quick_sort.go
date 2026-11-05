package l08sorting

import "math/rand/v2"

// QuickSort сортирует nums по возрастанию на месте быстрой сортировкой.
// Пакеты sort и slices не используй, смысл задачи в разбиении.
//
// Ожидаемая сложность: O(n log n) в среднем на любом входе, включая уже
// отсортированный, развёрнутый и состоящий из одинаковых чисел. Опорный
// элемент «всегда первый» и разбиение Ломуто, которое все равные кладёт
// в одну сторону, на таких входах вырождаются в O(n^2).
func QuickSort(nums []int) {
	for len(nums) > 1 {
		// случайный опорный: никакой фиксированный вход не станет худшим случаем
		p := nums[rand.IntN(len(nums))]
		lt, gt := partition3(nums, p)
		// рекурсия в меньшую часть, цикл по большей: глубина стека O(log n)
		if lt < len(nums)-gt {
			QuickSort(nums[:lt])
			nums = nums[gt:]
		} else {
			QuickSort(nums[gt:])
			nums = nums[:lt]
		}
	}
}

// partition3 раскладывает nums на три части: [0, lt) меньше p, [lt, gt)
// равны p, [gt, n) больше p (флаг Нидерландов, Дейкстра). Равные опорному
// сразу встают на место и в рекурсию не идут.
func partition3(nums []int, p int) (lt, gt int) {
	lt, i, gt := 0, 0, len(nums)
	for i < gt {
		switch {
		case nums[i] < p:
			nums[lt], nums[i] = nums[i], nums[lt]
			lt++
			i++
		case nums[i] > p:
			gt--
			nums[gt], nums[i] = nums[i], nums[gt]
		default:
			i++
		}
	}
	return lt, gt
}
