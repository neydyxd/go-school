package l03errors

// HTTPStatus переводит ошибку слоя бизнес-логики в HTTP-статус.
// Правила проверяются сверху вниз, срабатывает первое
// (ошибки могут быть обёрнуты в несколько слоёв и объединены через errors.Join):
//
//	nil                                     -> 200
//	errors.Is(err, ErrNotFound)             -> 404
//	внутри есть *FieldError (errors.As)     -> 400
//	errors.Is(err, context.DeadlineExceeded) -> 504
//	errors.Is(err, context.Canceled)         -> 499
//	всё остальное                           -> 500
func HTTPStatus(err error) int {
	panic("TODO: реализуй HTTPStatus")
}
