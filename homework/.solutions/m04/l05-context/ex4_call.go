package l05context

import (
	"context"
	"errors"
	"time"
)

// ErrSlow: причина отмены, когда Call не уложился в свой таймаут.
var ErrSlow = errors.New("l05context: операция не уложилась в таймаут")

// Call выполняет fn синхронно с таймаутом d и причиной ErrSlow
// (подсказка: context.WithTimeoutCause). fn обязана уважать контекст.
//
//   - fn вернула nil: вернуть nil, даже если таймаут сработал сразу после.
//   - fn вернула ошибку, а контекст fn к этому моменту отменён: вернуть
//     context.Cause этого контекста. Так наружу уходит настоящая причина:
//     ErrSlow для своего таймаута, причина родителя, если отменили его
//     (например, через context.WithCancelCause), DeadlineExceeded, если у
//     родителя дедлайн короче d.
//   - Иначе вернуть ошибку fn как есть.
func Call(ctx context.Context, d time.Duration, fn func(context.Context) error) error {
	ctx, cancel := context.WithTimeoutCause(ctx, d, ErrSlow)
	defer cancel()
	err := fn(ctx)
	if err == nil {
		return nil
	}
	if ctx.Err() != nil {
		return context.Cause(ctx)
	}
	return err
}
