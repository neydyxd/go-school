package l07godb

import (
	"context"
	"errors"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// MaxAttempts это сколько раз WithRetry запускает транзакцию, прежде чем сдаться.
const MaxAttempts = 5

// WithRetry выполняет fn внутри транзакции уровня SERIALIZABLE и коммитит её.
//
//   - если fn или COMMIT вернули ошибку PostgreSQL с кодом 40001
//     (serialization_failure) или 40P01 (deadlock_detected), транзакция
//     откатывается и запускается заново, всего не больше MaxAttempts попыток.
//     Ошибка может прийти обёрнутой (fmt.Errorf("...: %w", err)), ищи код
//     через errors.As(err, &pgErr) с pgErr типа *pgconn.PgError;
//   - попытки кончились: верни последнюю ошибку (так, чтобы errors.As её находил);
//   - любая другая ошибка fn: откат и сразу вернуть её, без повторов;
//   - ctx отменён: вернуть ошибку контекста, новых попыток не начинать.
//
// Между попытками можно (но не обязательно) подождать немного со случайным разбросом.
func WithRetry(ctx context.Context, pool *pgxpool.Pool, fn func(tx pgx.Tx) error) error {
	var err error
	for attempt := range MaxAttempts {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		// BeginTxFunc сам делает Rollback, если fn вернула ошибку, и Commit, если nil.
		err = pgx.BeginTxFunc(ctx, pool, pgx.TxOptions{IsoLevel: pgx.Serializable}, fn)
		if !retryable(err) {
			return err
		}
		// Короткая пауза с разбросом, чтобы конкуренты не столкнулись снова.
		pause := time.Duration(attempt+1)*5*time.Millisecond + rand.N(5*time.Millisecond)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(pause):
		}
	}
	return err
}

func retryable(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && (pgErr.Code == "40001" || pgErr.Code == "40P01")
}
