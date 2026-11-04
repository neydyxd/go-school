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
	panic("TODO: реализуй CollectUntil")
}
