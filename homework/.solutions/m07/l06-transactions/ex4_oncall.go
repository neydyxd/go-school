package l06transactions

import (
	"context"
	"errors"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDoctorNotFound = errors.New("врач не найден")
	ErrLastDoctor     = errors.New("нельзя уйти: это последний дежурный врач")
)

// GoOffCall снимает врача doctorID с дежурства (doctors.on_call = false).
// Инвариант: в больнице всегда остаётся хотя бы один дежурный врач.
//
//   - врача нет: ErrDoctorNotFound;
//   - врач уже не дежурит: ничего не делать, вернуть nil;
//   - врач последний дежурный: ErrLastDoctor, ничего не менять.
//
// Ловушка: два врача одновременно проверяют «дежурных двое, мне можно»
// и оба уходят. Это write skew: каждая транзакция меняет СВОЮ строку,
// поэтому блокировки строки врача недостаточно. Защитись одним из способов:
// заблокируй всех дежурных (SELECT ... FOR UPDATE) перед проверкой
// или работай в SERIALIZABLE с повтором при ошибке 40001.
func GoOffCall(ctx context.Context, pool *pgxpool.Pool, doctorID int64) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Блокируем всех дежурных. Конкурент, который уже снял кого-то
	// с дежурства, заставит нас подождать, а после его COMMIT Read Committed
	// перепроверит условие on_call и не вернёт ушедшего врача.
	rows, _ := tx.Query(ctx, `SELECT id FROM doctors WHERE on_call ORDER BY id FOR UPDATE`)
	onCall, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return err
	}
	if !slices.Contains(onCall, doctorID) {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM doctors WHERE id = $1)`, doctorID).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return ErrDoctorNotFound
		}
		return nil
	}
	if len(onCall) == 1 {
		return ErrLastDoctor
	}
	if _, err := tx.Exec(ctx, `UPDATE doctors SET on_call = false WHERE id = $1`, doctorID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
