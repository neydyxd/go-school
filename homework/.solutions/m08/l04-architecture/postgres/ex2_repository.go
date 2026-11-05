// Package postgres это слой хранения: SQL и перевод ошибок драйвера в доменные.
package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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
	err := r.pool.QueryRow(ctx,
		`INSERT INTO books (title, author, isbn, copies) VALUES ($1, $2, $3, $4) RETURNING id`,
		b.Title, b.Author, b.ISBN, b.Copies).Scan(&b.ID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return library.Book{}, library.ErrISBNTaken
	}
	if err != nil {
		return library.Book{}, fmt.Errorf("insert book: %w", err)
	}
	return b, nil
}

// Get читает книгу по ID. pgx.ErrNoRows это library.ErrNotFound.
func (r *BookRepo) Get(ctx context.Context, id int64) (library.Book, error) {
	return scanBook(r.pool.QueryRow(ctx, `SELECT `+columns+` FROM books WHERE id = $1`, id))
}

// TakeCopy уменьшает copies на 1 одним атомарным запросом
// (UPDATE ... SET copies = copies - 1 WHERE id = $1 AND copies > 0 RETURNING ...),
// а не «прочитать в Go, вычесть, записать»: два параллельных запроса выдадут один экземпляр дважды.
// Если UPDATE не нашёл строку, отличи «нет книги» (library.ErrNotFound)
// от «нет экземпляров» (library.ErrNoCopies) вторым запросом.
func (r *BookRepo) TakeCopy(ctx context.Context, id int64) (library.Book, error) {
	b, err := scanBook(r.pool.QueryRow(ctx,
		`UPDATE books SET copies = copies - 1 WHERE id = $1 AND copies > 0 RETURNING `+columns, id))
	if !errors.Is(err, library.ErrNotFound) {
		return b, err
	}
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM books WHERE id = $1)`, id).Scan(&exists); err != nil {
		return library.Book{}, fmt.Errorf("check book: %w", err)
	}
	if exists {
		return library.Book{}, library.ErrNoCopies
	}
	return library.Book{}, library.ErrNotFound
}

// uniqueViolation это код ошибки PostgreSQL для нарушения UNIQUE (Appendix A документации).
const uniqueViolation = "23505"

const columns = `id, title, author, isbn, copies`

func scanBook(row pgx.Row) (library.Book, error) {
	var b library.Book
	err := row.Scan(&b.ID, &b.Title, &b.Author, &b.ISBN, &b.Copies)
	if errors.Is(err, pgx.ErrNoRows) {
		return library.Book{}, library.ErrNotFound
	}
	if err != nil {
		return library.Book{}, fmt.Errorf("scan book: %w", err)
	}
	return b, nil
}
