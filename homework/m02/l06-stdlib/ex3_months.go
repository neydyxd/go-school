package l06stdlib

import "time"

// AddMonthsClamp прибавляет к t months календарных месяцев (может быть
// отрицательным). Если в целевом месяце нет такого числа, берётся последний
// день месяца: 31 января + 1 месяц = 28 (или 29) февраля, а не 3 марта,
// как сделал бы t.AddDate(0, 1, 0).
// Время суток и часовой пояс (Location) сохраняются.
func AddMonthsClamp(t time.Time, months int) time.Time {
	panic("TODO: реализуй AddMonthsClamp")
}
