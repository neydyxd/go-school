package l03select

import "time"

// CollectUntil читает значения из in и возвращает их, когда случится
// первое из трёх событий:
//   - прочитано limit значений (limit <= 0 значит «без лимита»);
//   - in закрыт;
//   - прошло timeout с момента вызова. Это ОБЩИЙ дедлайн на всю
//     функцию, а не пауза между значениями.
//
// Если значения приходят чаще, чем раз в timeout, функция всё равно
// обязана вернуться через timeout. Таймер после выхода останови.
func CollectUntil(in <-chan int, limit int, timeout time.Duration) []int {
	// Один таймер на весь вызов, создан ДО цикла.
	deadline := time.NewTimer(timeout)
	defer deadline.Stop()

	var out []int
	for limit <= 0 || len(out) < limit {
		select {
		case v, ok := <-in:
			if !ok {
				return out
			}
			out = append(out, v)
		case <-deadline.C:
			return out
		}
	}
	return out
}
