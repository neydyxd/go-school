package repository

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"goschool/tasktracker/internal/service"
	"goschool/tasktracker/migrations"
)

// Образец интеграционного теста репозитория в настоящей базе.
// Каждый тест получает свою схему, в которую накатываются миграции:
// тесты изолированы и могут идти параллельно. Тот же приём, что в
// homework/internal/testdb. База: TEST_DATABASE_URL или учебная из homework
// (cd homework && docker compose up -d).
func newTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	base := os.Getenv("TEST_DATABASE_URL")
	if base == "" {
		base = "postgres://goschool:goschool@localhost:55432/goschool?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	admin, err := pgxpool.New(ctx, base)
	if err != nil {
		t.Fatalf("база недоступна: %v", err)
	}
	defer admin.Close()
	b := make([]byte, 6)
	_, _ = rand.Read(b)
	schema := "tt_" + hex.EncodeToString(b)
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("база недоступна (cd homework && docker compose up -d): %v", err)
	}
	t.Cleanup(func() {
		p, err := pgxpool.New(context.Background(), base)
		if err == nil {
			_, _ = p.Exec(context.Background(), "DROP SCHEMA "+schema+" CASCADE")
			p.Close()
		}
	})

	u, _ := url.Parse(base)
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	pool, err := pgxpool.New(ctx, u.String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	if err := migrations.Up(ctx, db); err != nil {
		t.Fatalf("миграции: %v", err)
	}
	return pool
}

func TestUsersCreate(t *testing.T) {
	ctx := context.Background()
	r := NewUsers(newTestPool(t))

	u, err := r.Create(ctx, service.User{Email: "a@example.com", PasswordHash: "h", Role: "user"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if u.ID == 0 || u.CreatedAt.IsZero() {
		t.Errorf("Create = %+v: want ID и CreatedAt из базы", u)
	}
	_, err = r.Create(ctx, service.User{Email: "a@example.com", PasswordHash: "h2", Role: "user"})
	if !errors.Is(err, service.ErrConflict) {
		t.Errorf("повторный email: err = %v, want service.ErrConflict", err)
	}
}

// Образец: несколько записей в одной транзакции через TxManager.
// В модуле 09 так же в одну транзакцию попадёт постановка задачи River (InsertTx).
func TestTxManagerAtomic(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	tm := NewTxManager(pool)
	count := func(email string) int {
		var n int
		if err := pool.QueryRow(ctx, `SELECT count(*) FROM users WHERE email = $1`, email).Scan(&n); err != nil {
			t.Fatal(err)
		}
		return n
	}
	user := func(email string) service.User {
		return service.User{Email: email, PasswordHash: "h", Role: "user"}
	}

	if err := tm.InTx(ctx, func(s service.Store) error {
		if _, err := s.Users.Create(ctx, user("a@example.com")); err != nil {
			return err
		}
		_, err := s.Users.Create(ctx, user("b@example.com"))
		return err
	}); err != nil {
		t.Fatalf("InTx: %v", err)
	}
	if count("a@example.com") != 1 || count("b@example.com") != 1 {
		t.Fatal("после успешной транзакции обе записи должны быть в базе")
	}

	// вторая запись падает (email занят): первая тоже должна откатиться
	err := tm.InTx(ctx, func(s service.Store) error {
		if _, err := s.Users.Create(ctx, user("c@example.com")); err != nil {
			return err
		}
		_, err := s.Users.Create(ctx, user("a@example.com"))
		return err
	})
	if !errors.Is(err, service.ErrConflict) {
		t.Fatalf("InTx err = %v, want service.ErrConflict", err)
	}
	if count("c@example.com") != 0 {
		t.Error("c@example.com осталась в базе: вставки не в одной транзакции")
	}

	// ошибка бизнес-логики после записи тоже откатывает запись
	boom := errors.New("boom")
	if err := tm.InTx(ctx, func(s service.Store) error {
		if _, err := s.Users.Create(ctx, user("d@example.com")); err != nil {
			return err
		}
		return boom
	}); !errors.Is(err, boom) {
		t.Fatalf("InTx err = %v, want boom", err)
	}
	if count("d@example.com") != 0 {
		t.Error("d@example.com осталась в базе после ошибки в fn")
	}
}
