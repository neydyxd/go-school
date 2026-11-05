package l06binarysearch

// ShipWithinDays: на складе очередь посылок с весами weights, их надо
// отгрузить строго по порядку за days дней. Каждый день грузовик берёт
// подряд идущие посылки, пока их суммарный вес не превышает его
// грузоподъёмность. Верни минимальную грузоподъёмность, при которой
// успеваем за days дней. Веса положительные, weights не пустой, days >= 1.
// [1 2 3 4 5 6 7 8 9 10] за 5 дней даёт 15.
//
// Ожидаемая сложность: O(n log(сумма весов)). Перебор грузоподъёмности
// по одной единице слишком медленный.
func ShipWithinDays(weights []int, days int) int {
	lo, hi := 0, 0 // ответ лежит в [max(weights), sum(weights)]
	for _, w := range weights {
		lo = max(lo, w)
		hi += w
	}
	for lo < hi {
		mid := lo + (hi-lo)/2
		if daysNeeded(weights, mid) <= days {
			hi = mid // mid хватает, пробуем меньше
		} else {
			lo = mid + 1
		}
	}
	return lo
}

// daysNeeded жадно грузит посылки и считает дни при грузоподъёмности c.
func daysNeeded(weights []int, c int) int {
	d, load := 1, 0
	for _, w := range weights {
		if load+w > c {
			d++
			load = 0
		}
		load += w
	}
	return d
}
