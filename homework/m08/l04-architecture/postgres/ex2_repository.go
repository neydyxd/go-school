// Package postgres это слой хранения: SQL и перевод ошибок драйвера в доменные.
package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/homework/m08/l04-architecture/library"
)

// BookRepo хранит книги в таблице books (схема в testdata/schema.sql).
// Он удовлетворяет service.BookRepository, но пакет service не импортирует:
// интерфейс в Go реализуется неявно.
type BookRepo struct {
	pool *pgxpool.Pool
}

// NewBookRepo создаёт репозиторий поверх пула.
func NewBookRepo(pool *pgxpool.Pool) *BookRepo {
	return &BookRepo{pool: pool}
}

// Create вставляет книгу (INSERT ... RETURNING id) и возвращает её с ID.
// Нарушение UNIQUE по isbn (код ошибки PostgreSQL 23505, *pgconn.PgError) это library.ErrISBNTaken.
// Не сравнивай текст ошибки со строкой "duplicate key": он зависит от языка сервера.
func (r *BookRepo) Create(ctx context.Context, b library.Book) (library.Book, error) {
	panic("TODO: реализуй BookRepo.Create")
}

// Get читает книгу по ID. pgx.ErrNoRows это library.ErrNotFound.
func (r *BookRepo) Get(ctx context.Context, id int64) (library.Book, error) {
	panic("TODO: реализуй BookRepo.Get")
}

// TakeCopy уменьшает copies на 1 одним атомарным запросом
// (UPDATE ... SET copies = copies - 1 WHERE id = $1 AND copies > 0 RETURNING ...),
// а не «прочитать в Go, вычесть, записать»: два параллельных запроса выдадут один экземпляр дважды.
// Если UPDATE не нашёл строку, отличи «нет книги» (library.ErrNotFound)
// от «нет экземпляров» (library.ErrNoCopies) вторым запросом.
func (r *BookRepo) TakeCopy(ctx context.Context, id int64) (library.Book, error) {
	panic("TODO: реализуй BookRepo.TakeCopy")
}
