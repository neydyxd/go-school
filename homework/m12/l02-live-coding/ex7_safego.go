package l02livecoding

// PanicError: паника, перехваченная в горутине и превращённая в ошибку.
type PanicError struct {
	Value any    // то, что вернул recover()
	Stack []byte // стек в момент паники (runtime/debug.Stack)
}

// Error возвращает "panic: <Value>" (Value через fmt %v).
func (e *PanicError) Error() string {
	panic("TODO: реализуй Error")
}

// Unwrap возвращает Value, если это error, иначе nil. Так errors.Is и errors.As
// видят ошибку, с которой запаниковали (в том числе runtime.Error).
func (e *PanicError) Unwrap() error {
	panic("TODO: реализуй Unwrap")
}

// RunAll запускает каждую функцию в своей горутине, ждёт все и возвращает
// ошибки в порядке fns: errs[i] это ошибка fns[i] или *PanicError, если fns[i]
// запаниковала. Паника одной функции не роняет программу и не мешает остальным.
// Без функций: пустой слайс.
func RunAll(fns ...func() error) []error {
	panic("TODO: реализуй RunAll")
}
