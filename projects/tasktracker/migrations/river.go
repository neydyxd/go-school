package migrations

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
)

// RiverUp создаёт или обновляет таблицы River (river_job, river_leader,
// river_queue и другие) в схеме из search_path. Готово.
//
// Миграции River живут отдельно от goose: их версии ведёт сам River в таблице
// river_migration. Альтернатива для прода: river migrate-get выгружает SQL,
// и его кладут отдельным файлом goose (см. урок m09/l02).
func RiverUp(ctx context.Context, pool *pgxpool.Pool) error {
	m, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		return fmt.Errorf("rivermigrate: %w", err)
	}
	if _, err := m.Migrate(ctx, rivermigrate.DirectionUp, nil); err != nil {
		return fmt.Errorf("river migrate up: %w", err)
	}
	return nil
}
