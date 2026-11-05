// Package migrations встраивает SQL-миграции goose в бинарник: образ сам умеет
// мигрировать базу (`/app -migrate`), отдельный goose в контейнере не нужен.
package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var fs embed.FS

// Up применяет все новые миграции и возвращает, сколько применено.
func Up(ctx context.Context, db *sql.DB) (int, error) {
	p, err := goose.NewProvider(goose.DialectPostgres, db, fs)
	if err != nil {
		return 0, fmt.Errorf("goose: %w", err)
	}
	res, err := p.Up(ctx)
	if err != nil {
		return 0, fmt.Errorf("goose up: %w", err)
	}
	return len(res), nil
}
