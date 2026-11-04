package l03functions

// NewCounter возвращает две функции, разделяющие одно состояние:
//   - next увеличивает счётчик на step и возвращает новое значение:
//     для step = 2 вызовы next дают 2, 4, 6, ...
//   - reset сбрасывает счётчик в ноль, следующий next снова вернёт step.
//
// Разные вызовы NewCounter дают независимые счётчики.
func NewCounter(step int) (next func() int, reset func()) {
	n := 0 // обе функции захватывают одну и ту же переменную
	next = func() int {
		n += step
		return n
	}
	reset = func() { n = 0 }
	return next, reset
}
