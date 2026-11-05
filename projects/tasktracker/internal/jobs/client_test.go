package jobs

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Образец: конфиг клиента валиден (очереди, воркеры, периодическая задача).
// База не нужна: pgxpool.New не подключается, пока пул не попросят о соединении,
// а NewClient только проверяет конфигурацию.
func TestNewClient(t *testing.T) {
	pool, err := pgxpool.New(context.Background(), "postgres://nobody@127.0.0.1:1/none")
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if _, err := NewClient(pool, slog.New(slog.NewTextHandler(io.Discard, nil)), &FakeMailer{}); err != nil {
		t.Fatalf("NewClient: %v", err)
	}
}
