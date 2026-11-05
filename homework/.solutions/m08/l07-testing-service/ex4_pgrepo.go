package l07testingservice

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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

const taskCols = `id, owner_id, title, status, created_at`

func scanTask(row pgx.Row) (Task, error) {
	var t Task
	err := row.Scan(&t.ID, &t.OwnerID, &t.Title, &t.Status, &t.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	return t, err
}

func (r *PgRepo) Create(ctx context.Context, t Task) (Task, error) {
	return scanTask(r.pool.QueryRow(ctx,
		`INSERT INTO tasks (owner_id, title, status)
		 VALUES ($1, $2, COALESCE(NULLIF($3, ''), 'todo'))
		 RETURNING `+taskCols,
		t.OwnerID, t.Title, t.Status))
}

func (r *PgRepo) Get(ctx context.Context, ownerID, id int64) (Task, error) {
	return scanTask(r.pool.QueryRow(ctx,
		`SELECT `+taskCols+` FROM tasks WHERE owner_id = $1 AND id = $2`, ownerID, id))
}

func (r *PgRepo) List(ctx context.Context, ownerID int64, f Filter) ([]Task, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+taskCols+` FROM tasks
		 WHERE owner_id = $1 AND id > $2 AND ($3 = '' OR status = $3)
		 ORDER BY id
		 LIMIT $4`,
		ownerID, f.AfterID, f.Status, f.PageSize())
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, func(row pgx.CollectableRow) (Task, error) { return scanTask(row) })
}

func (r *PgRepo) UpdateStatus(ctx context.Context, ownerID, id int64, status string) (Task, error) {
	return scanTask(r.pool.QueryRow(ctx,
		`UPDATE tasks SET status = $3 WHERE owner_id = $1 AND id = $2 RETURNING `+taskCols,
		ownerID, id, status))
}
