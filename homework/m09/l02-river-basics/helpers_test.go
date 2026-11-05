package l02riverbasics

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
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

// insertOnlyClient: клиент без очередей, только для вставки задач.
func insertOnlyClient(t *testing.T, pool *pgxpool.Pool) *river.Client[pgx.Tx] {
	t.Helper()
	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		t.Fatal(err)
	}
	return client
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

func addUser(t *testing.T, pool *pgxpool.Pool, email string) int64 {
	t.Helper()
	var id int64
	if err := pool.QueryRow(testCtx(t), `INSERT INTO users (email) VALUES ($1) RETURNING id`, email).Scan(&id); err != nil {
		t.Fatal(err)
	}
	return id
}

// fakeMailer запоминает письма; failNext раз подряд отвечает ошибкой.
type fakeMailer struct {
	mu       sync.Mutex
	sent     []string // "to|subject"
	failNext int
}

var errSMTP = errors.New("SMTP: 451 попробуйте позже")

func (m *fakeMailer) Send(_ context.Context, to, subject, _ string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.failNext > 0 {
		m.failNext--
		return errSMTP
	}
	m.sent = append(m.sent, fmt.Sprintf("%s|%s", to, subject))
	return nil
}

func (m *fakeMailer) Sent() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.sent...)
}
