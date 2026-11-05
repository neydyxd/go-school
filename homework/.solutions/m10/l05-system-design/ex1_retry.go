package l05systemdesign

import (
	"context"
	"errors"
	"fmt"
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
	attempts := max(p.Attempts, 1)
	sleep := p.Sleep
	if sleep == nil {
		sleep = sleepCtx
	}
	var last error
	for i := range attempts {
		if err := ctx.Err(); err != nil {
			return interrupted(err, last)
		}
		last = fn(ctx)
		if last == nil {
			return nil
		}
		if errors.Is(last, ErrPermanent) {
			return last
		}
		if i == attempts-1 {
			break
		}
		d := backoff(p.Base, p.Max, i)
		if p.Jitter != nil {
			d = p.Jitter(d)
		}
		if err := sleep(ctx, d); err != nil {
			return interrupted(err, last)
		}
	}
	return fmt.Errorf("после %d попыток: %w", attempts, last)
}

// backoff: min(maxD, base * 2^i) без переполнения.
func backoff(base, maxD time.Duration, i int) time.Duration {
	d := base
	for range i {
		if d >= maxD/2 {
			return maxD
		}
		d *= 2
	}
	return min(d, maxD)
}

func interrupted(ctxErr, last error) error {
	if last == nil {
		return fmt.Errorf("retry прерван: %w", ctxErr)
	}
	return fmt.Errorf("retry прерван: %w (последняя ошибка: %w)", ctxErr, last)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
