package l01whyqueues

import (
	"context"
	"time"
)

// PermanentError помечает ошибку, которую повторять бессмысленно: 400 от
// партнёра, «пользователь удалён», невалидные данные. Готово.
type PermanentError struct{ Err error }

func (e *PermanentError) Error() string { return "permanent: " + e.Err.Error() }
func (e *PermanentError) Unwrap() error { return e.Err }

// Permanent заворачивает err в *PermanentError. Готово.
func Permanent(err error) error { return &PermanentError{Err: err} }

// Retry вызывает fn, пока она не вернёт nil, но не больше maxAttempts раз
// всего (первая попытка плюс повторы). maxAttempts < 1 считай за 1.
//
// Между попытками ждёт wait(n), где n это номер ПОВТОРА: перед второй попыткой
// wait(1), перед третьей wait(2). После последней попытки не ждёт.
// Ожидание прерывается отменой ctx.
//
// Возвращает:
//   - nil, как только fn вернула nil;
//   - ошибку fn сразу, без повторов, если это *PermanentError (errors.As);
//   - последнюю ошибку fn, если попытки кончились;
//   - ctx.Err(), если ctx отменён до попытки или во время ожидания
//     (fn при отменённом ctx не вызывай).
func Retry(ctx context.Context, maxAttempts int, wait func(retry int) time.Duration, fn func(ctx context.Context) error) error {
	panic("TODO: Retry")
}
