package l07godb

import (
	"context"

	"github.com/jackc/pgx/v5"
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
	panic("TODO: WithRetry")
}
