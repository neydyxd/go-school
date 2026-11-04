package l03testabledesign

import "time"

// Clock это зависимость от времени. В проде её реализует realClock,
// а в тестах подставляется фейк с заранее заданным временем.
type Clock interface {
	Now() time.Time
}

type realClock struct{}

func (realClock) Now() time.Time { return time.Now() }

// SystemClock это настоящие часы для продакшен-кода.
var SystemClock Clock = realClock{}

// Greeting возвращает приветствие по часу из c.Now():
//   - 05:00-11:59 "Доброе утро"
//   - 12:00-17:59 "Добрый день"
//   - 18:00-22:59 "Добрый вечер"
//   - 23:00-04:59 "Доброй ночи"
//
// Час бери в том часовом поясе, в котором пришло время (t.Hour()), без перевода в UTC.
func Greeting(c Clock) string {
	switch h := c.Now().Hour(); {
	case h >= 5 && h < 12:
		return "Доброе утро"
	case h >= 12 && h < 18:
		return "Добрый день"
	case h >= 18 && h < 23:
		return "Добрый вечер"
	default:
		return "Доброй ночи"
	}
}
