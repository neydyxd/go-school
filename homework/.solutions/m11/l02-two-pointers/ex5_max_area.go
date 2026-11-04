package l02twopointers

// MaxArea получает высоты вертикальных отрезков, стоящих в точках 0, 1, 2...
// Любые два отрезка i < j вместе с осью образуют сосуд с площадью
// (j - i) * min(heights[i], heights[j]). Верни максимальную площадь.
// Меньше двух отрезков: 0.
//
// Ожидаемая сложность: O(n) по времени.
func MaxArea(heights []int) int {
	best := 0
	l, r := 0, len(heights)-1
	for l < r {
		h := min(heights[l], heights[r])
		best = max(best, (r-l)*h)
		// Двигаем более низкий борт: с ним площадь уже не вырастет,
		// ведь ширина дальше только уменьшается.
		if heights[l] < heights[r] {
			l++
		} else {
			r--
		}
	}
	return best
}
