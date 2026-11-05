package l01whyqueues

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/homework/internal/testdb"
)

func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func newPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testdb.New(t, "testdata/seed.sql")
}
