package l07godb

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrEmailTaken = errors.New("email уже занят")

// CreateUser добавляет пользователя и возвращает его id (RETURNING id).
//   - phone == nil записывается как NULL;
//   - email уже есть: ErrEmailTaken. Не делай предварительный SELECT
//     («проверить, потом вставить» гоняется с параллельной регистрацией):
//     вставляй сразу и распознавай ошибку уникальности *pgconn.PgError
//     с кодом "23505" (unique_violation) и именем ограничения users_email_key;
//   - остальные ошибки возвращай как есть, обернув через %w.
func (r *Repo) CreateUser(ctx context.Context, email, name string, phone *string) (int64, error) {
	var id int64
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (email, name, phone) VALUES ($1, $2, $3) RETURNING id`,
		email, name, phone).Scan(&id)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_key" {
		return 0, ErrEmailTaken
	}
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	return id, nil
}
