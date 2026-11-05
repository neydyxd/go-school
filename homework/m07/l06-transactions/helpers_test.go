package l06transactions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/homework/internal/testdb"
)

func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func newDB(t *testing.T) *pgxpool.Pool {
	t.Helper()
	return testdb.New(t, "testdata/seed.sql")
}

// holder это «соседняя» транзакция на отдельном соединении: тест открывает её,
// делает изменения и держит блокировки, пока сам не решит закоммитить.
type holder struct {
	conn *pgxpool.Conn
	tx   pgx.Tx
	pid  uint32
}

func hold(t *testing.T, pool *pgxpool.Pool, iso pgx.TxIsoLevel, stmts ...string) *holder {
	t.Helper()
	ctx := testCtx(t)
	conn, err := pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("не взял соединение: %v", err)
	}
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{IsoLevel: iso})
	if err != nil {
		t.Fatalf("BEGIN: %v", err)
	}
	h := &holder{conn: conn, tx: tx, pid: conn.Conn().PgConn().PID()}
	t.Cleanup(func() {
		_ = h.tx.Rollback(context.Background())
		h.conn.Release()
	})
	for _, s := range stmts {
		if _, err := tx.Exec(ctx, s); err != nil {
			t.Fatalf("соседняя транзакция: %v\n%s", err, s)
		}
	}
	return h
}

func (h *holder) commit() error {
	return h.tx.Commit(context.Background())
}

// call запускает функцию студента в отдельной горутине.
type call struct {
	done chan struct{}
	err  error
}

func start(f func() error) *call {
	c := &call{done: make(chan struct{})}
	go func() {
		defer close(c.done)
		c.err = f()
	}()
	return c
}

func (c *call) wait(t *testing.T, what string) error {
	t.Helper()
	select {
	case <-c.done:
		return c.err
	case <-time.After(10 * time.Second):
		t.Fatalf("%s не завершилась за 10 с: похоже на вечное ожидание блокировки", what)
		return nil
	}
}

// waitBlockedOrDone ждёт одно из двух событий: функция студента встала
// в очередь за блокировкой транзакции h (true) или уже завершилась (false).
// Проверяем через pg_blocking_pids, а не через sleep: тест детерминирован.
func waitBlockedOrDone(t *testing.T, pool *pgxpool.Pool, h *holder, c *call) bool {
	t.Helper()
	deadline := time.After(10 * time.Second)
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-c.done:
			return false
		case <-deadline:
			t.Fatalf("за 10 с функция не завершилась и не встала в ожидание блокировки")
		case <-tick.C:
		}
		var n int
		err := pool.QueryRow(testCtx(t),
			`SELECT count(*) FROM pg_stat_activity WHERE $1::int = ANY (pg_blocking_pids(pid))`,
			int32(h.pid)).Scan(&n)
		if err != nil {
			t.Fatalf("pg_stat_activity: %v", err)
		}
		if n > 0 {
			return true
		}
	}
}

func pgCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
