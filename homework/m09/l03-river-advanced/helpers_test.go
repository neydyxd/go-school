package l03riveradvanced

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"

	"goschool/homework/internal/testdb"
)

func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// newRiverPool: своя схема с таблицами урока и таблицами River.
func newRiverPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := testdb.New(t, "testdata/seed.sql")
	migrator, err := rivermigrate.New(riverpgxv5.New(pool), nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := migrator.Migrate(testCtx(t), rivermigrate.DirectionUp, nil); err != nil {
		t.Fatalf("миграции River: %v", err)
	}
	return pool
}

func beginTx(t *testing.T, pool *pgxpool.Pool) pgx.Tx {
	t.Helper()
	tx, err := pool.Begin(testCtx(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
	return tx
}

// near проверяет, что got отстоит от from на want с допуском 2 секунды.
func near(got, from time.Time, want time.Duration) bool {
	d := got.Sub(from) - want
	return d > -2*time.Second && d < 2*time.Second
}
