package l07godb

import (
	"context"
	"errors"
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
	panic("TODO: CreateUser")
}
