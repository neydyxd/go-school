package l07godb

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("не найдено")

// User это строка таблицы users. Теги db нужны для pgx.RowToStructByName.
// Phone указатель, потому что в базе телефон может быть NULL.
type User struct {
	ID    int64   `db:"id"`
	Email string  `db:"email"`
	Name  string  `db:"name"`
	Phone *string `db:"phone"`
}

// Repo это репозиторий поверх пула соединений.
type Repo struct {
	pool *pgxpool.Pool
}

func NewRepo(pool *pgxpool.Pool) *Repo { return &Repo{pool: pool} }

// UserByEmail возвращает пользователя с этим email (точное совпадение).
//   - NULL в phone превращается в Phone == nil, пустая строка в базе
//     остаётся указателем на "";
//   - пользователя нет: ErrNotFound (проверяется через errors.Is), а не pgx.ErrNoRows:
//     вызывающий код не должен знать, какой драйвер внутри.
func (r *Repo) UserByEmail(ctx context.Context, email string) (User, error) {
	rows, _ := r.pool.Query(ctx, `SELECT id, email, name, phone FROM users WHERE email = $1`, email)
	u, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[User])
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("user by email: %w", err)
	}
	return u, nil
}
