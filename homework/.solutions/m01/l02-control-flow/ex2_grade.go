package l02controlflow

// Grade переводит баллы в оценку:
//
//	90-100 "отлично"
//	75-89  "хорошо"
//	60-74  "удовлетворительно"
//	0-59   "неудовлетворительно"
//
// Баллы вне диапазона 0-100 дают "ошибка".
// Попробуй написать через switch без выражения (switch { case ...: }).
func Grade(score int) string {
	switch {
	case score < 0 || score > 100:
		return "ошибка"
	case score >= 90:
		return "отлично"
	case score >= 75:
		return "хорошо"
	case score >= 60:
		return "удовлетворительно"
	default:
		return "неудовлетворительно"
	}
}
