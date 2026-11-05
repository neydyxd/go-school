package l07testingservice

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PgRepo это TaskRepo поверх PostgreSQL (таблица tasks из testdata/schema.sql).
// Он обязан пройти тот же контрактный тест, что и FakeRepo.
//
// Подсказки:
//   - pgx.ErrNoRows превращай в ErrNotFound (errors.Is);
//   - владельца проверяй в самом запросе (WHERE owner_id = $1 AND id = $2),
//     а не отдельным SELECT: так нет гонки и лишнего похода в базу;
//   - created_at и статус по умолчанию ставит база, забирай их через RETURNING;
//   - только плейсхолдеры $1, $2, никакой склейки строк с данными.
type PgRepo struct {
	pool *pgxpool.Pool
}

// NewPgRepo готов, менять не нужно.
func NewPgRepo(pool *pgxpool.Pool) *PgRepo { return &PgRepo{pool: pool} }

func (r *PgRepo) Create(ctx context.Context, t Task) (Task, error) {
	panic("TODO: реализуй Create")
}

func (r *PgRepo) Get(ctx context.Context, ownerID, id int64) (Task, error) {
	panic("TODO: реализуй Get")
}

func (r *PgRepo) List(ctx context.Context, ownerID int64, f Filter) ([]Task, error) {
	panic("TODO: реализуй List")
}

func (r *PgRepo) UpdateStatus(ctx context.Context, ownerID, id int64, status string) (Task, error) {
	panic("TODO: реализуй UpdateStatus")
}
