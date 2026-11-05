package l06transactions

import (
	"context"
	"errors"

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
	panic("TODO: GoOffCall")
}
