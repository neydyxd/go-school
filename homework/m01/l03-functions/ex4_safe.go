package l03functions

// Safe вызывает f и превращает панику внутри неё в обычную ошибку.
//   - f отработала без паники: nil;
//   - паника со значением-ошибкой (panic(err)): ошибка с текстом "panic: " + err.Error(),
//     причём errors.Is(result, err) должна быть true (оберни через %w);
//   - паника с любым другим значением: ошибка с текстом "panic: " + fmt.Sprint(значение);
//   - panic(nil) тоже паника: результат не nil.
//
// Подсказка: именованный результат, defer и recover.
func Safe(f func()) (err error) {
	panic("TODO: реализуй Safe")
}
