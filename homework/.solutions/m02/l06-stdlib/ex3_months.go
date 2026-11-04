package l06stdlib

import "time"

func AddMonthsClamp(t time.Time, months int) time.Time {
	y, m, d := t.Date()
	// Первое число целевого месяца: time.Date нормализует месяц вне 1..12.
	first := time.Date(y, m+time.Month(months), 1, 0, 0, 0, 0, t.Location())
	// Нулевой день следующего месяца это последний день текущего.
	last := time.Date(first.Year(), first.Month()+1, 0, 0, 0, 0, 0, t.Location()).Day()
	d = min(d, last)
	hh, mm, ss := t.Clock()
	return time.Date(first.Year(), first.Month(), d, hh, mm, ss, t.Nanosecond(), t.Location())
}
