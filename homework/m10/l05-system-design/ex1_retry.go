package l05systemdesign

import (
	"context"
	"errors"
	"time"
)

// ErrPermanent помечает ошибку, которую повторять бессмысленно (400, 404,
// неверный пароль). fn оборачивает её: fmt.Errorf("...: %w", ErrPermanent).
var ErrPermanent = errors.New("l05systemdesign: постоянная ошибка")

// RetryPolicy: сколько раз и с какими паузами повторять.
type RetryPolicy struct {
	Attempts int           // всего попыток, включая первую; меньше 1 считается как 1
	Base     time.Duration // пауза после первой неудачи (до jitter)
	Max      time.Duration // потолок паузы (до jitter)

	// Jitter получает паузу d и возвращает случайную в [0, d] (full jitter).
	// nil: без jitter. В тестах подставляется детерминированная функция.
	Jitter func(d time.Duration) time.Duration
	// Sleep ждёт d или отмены ctx (тогда возвращает ошибку контекста).
	// nil: настоящий сон через time.Timer с учётом ctx.
	Sleep func(ctx context.Context, d time.Duration) error
}

// Retry вызывает fn, пока она не вернёт nil, но не больше p.Attempts раз.
//
//   - Пауза после неудачной попытки номер i (с нуля): min(Max, Base * 2^i),
//     затем через Jitter, если он задан. Умножение не должно переполняться
//     даже на сотой попытке: пауза всегда в пределах (0, Max].
//   - После последней попытки не спать.
//   - Ошибка с ErrPermanent (errors.Is): сразу вернуть её как есть, без пауз.
//   - Все попытки неудачны: вернуть ошибку, из которой errors.Is достаёт
//     последнюю ошибку fn, с текстом "после N попыток: ...".
//   - ctx отменён (до попытки или во время паузы): вернуть ошибку, из которой
//     errors.Is достаёт и ошибку контекста, и последнюю ошибку fn (если она была).
func Retry(ctx context.Context, p RetryPolicy, fn func(ctx context.Context) error) error {
	panic("TODO: реализуй Retry")
}
