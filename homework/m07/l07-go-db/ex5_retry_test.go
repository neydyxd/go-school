package l07godb

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"goschool/homework/internal/testdb"
)

// Настоящий конфликт сериализации: между чтением и записью счётчика
// соседнее соединение успевает его изменить и закоммитить.
func TestWithRetryRealConflict(t *testing.T) {
	t.Parallel()
	_, pool := newRepo(t)
	ctx := testCtx(t)
	attempts := 0
	err := WithRetry(ctx, pool, func(tx pgx.Tx) error {
		attempts++
		var v int64
		if err := tx.QueryRow(ctx, `SELECT value FROM counters WHERE name = 'hits'`).Scan(&v); err != nil {
			return err
		}
		if attempts == 1 {
			// соседняя транзакция на другом соединении из пула, autocommit
			if _, err := pool.Exec(ctx, `UPDATE counters SET value = value + 100 WHERE name = 'hits'`); err != nil {
				return fmt.Errorf("сосед: %w", err)
			}
		}
		_, err := tx.Exec(ctx, `UPDATE counters SET value = $1 WHERE name = 'hits'`, v+1)
		return err
	})
	if err != nil {
		t.Fatalf("WithRetry = %v (код %q), want nil: 40001 надо повторять", err, code(err))
	}
	if attempts != 2 {
		t.Errorf("fn вызвана %d раз, want 2", attempts)
	}
	if v := testdb.Rows(t, pool, `SELECT value FROM counters`); v[0][0] != "101" {
		t.Errorf("value = %s, want 101: соседское +100 потеряно, транзакция должна быть SERIALIZABLE", v[0][0])
	}
}

func code(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}

var errBusiness = errors.New("бизнес-ошибка")

func TestWithRetry(t *testing.T) {
	tests := []struct {
		name         string
		failures     int   // сколько первых вызовов вернут failErr
		failErr      error // ошибка, которую вернёт fn
		wantErr      error // nil или ошибка для errors.Is
		wantCode     string
		wantAttempts int
		wantRows     string // сколько строк hits-* осталось в counters
	}{
		{"успех с первого раза", 0, nil, nil, "", 1, "1"},
		{"две ошибки 40001, потом успех", 2, &pgconn.PgError{Code: "40001"}, nil, "", 3, "1"},
		{"обёрнутая 40001", 1, fmt.Errorf("обёртка: %w", &pgconn.PgError{Code: "40001"}), nil, "", 2, "1"},
		{"deadlock тоже повторяем", 1, &pgconn.PgError{Code: "40P01"}, nil, "", 2, "1"},
		{"попытки кончились", 100, fmt.Errorf("обёртка: %w", &pgconn.PgError{Code: "40001"}), nil, "40001", MaxAttempts, "0"},
		{"бизнес-ошибку не повторяем", 100, errBusiness, errBusiness, "", 1, "0"},
		{"нарушение уникальности не повторяем", 100, &pgconn.PgError{Code: "23505"}, nil, "23505", 1, "0"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, pool := newRepo(t)
			ctx := testCtx(t)
			attempts := 0
			err := WithRetry(ctx, pool, func(tx pgx.Tx) error {
				attempts++
				// пишем строку: она должна остаться только у закоммиченной попытки
				if _, err := tx.Exec(ctx, `INSERT INTO counters VALUES ($1, 1)`, fmt.Sprintf("hits-%d", attempts)); err != nil {
					return err
				}
				if attempts <= tt.failures {
					return tt.failErr
				}
				return nil
			})
			switch {
			case tt.wantErr != nil && !errors.Is(err, tt.wantErr):
				t.Fatalf("WithRetry = %v, want %v", err, tt.wantErr)
			case tt.wantCode != "" && code(err) != tt.wantCode:
				t.Fatalf("WithRetry = %v, want ошибку с кодом %s (через errors.As)", err, tt.wantCode)
			case tt.wantErr == nil && tt.wantCode == "" && err != nil:
				t.Fatalf("WithRetry = %v, want nil", err)
			}
			if attempts != tt.wantAttempts {
				t.Errorf("fn вызвана %d раз, want %d", attempts, tt.wantAttempts)
			}
			if n := testdb.Rows(t, pool, `SELECT count(*) FROM counters WHERE name LIKE 'hits-%'`); n[0][0] != tt.wantRows {
				t.Errorf("закоммичено попыток %s, want %s: неудачные попытки нужно откатывать", n[0][0], tt.wantRows)
			}
			if s := pool.Stat(); s.AcquiredConns() != 0 {
				t.Errorf("после WithRetry занято %d соединений, want 0", s.AcquiredConns())
			}
		})
	}
}

func TestWithRetryCanceled(t *testing.T) {
	t.Parallel()
	_, pool := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	err := WithRetry(ctx, pool, func(tx pgx.Tx) error {
		attempts++
		cancel() // пока шла первая попытка, клиент ушёл
		return &pgconn.PgError{Code: "40001"}
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WithRetry после отмены = %v, want context.Canceled", err)
	}
	if attempts != 1 {
		t.Errorf("fn вызвана %d раз после отмены контекста, want 1", attempts)
	}
}

// Конфликт, который всплывает только на COMMIT. Соседняя SERIALIZABLE-транзакция
// прочитала счётчики и поменяла строку a. fn читает те же счётчики и меняет
// строку b, а пока fn работает, сосед коммитится. Каждая прочитала то, что
// изменила другая (write skew), и Postgres отменяет fn на COMMIT с 40001.
// Ретраить нужно и такую ошибку, а не только ошибки самой fn.
func TestWithRetryCommitConflict(t *testing.T) {
	t.Parallel()
	_, pool := newRepo(t)
	ctx := testCtx(t)
	testdb.Exec(t, pool, `INSERT INTO counters VALUES ('a', 0), ('b', 0)`)
	h := holdIso(t, pool, pgx.Serializable,
		`SELECT sum(value) FROM counters`,
		`UPDATE counters SET value = value + 1 WHERE name = 'a'`)

	attempts := 0
	var fnErr error
	err := WithRetry(ctx, pool, func(tx pgx.Tx) error {
		attempts++
		var sum int64
		if err := tx.QueryRow(ctx, `SELECT sum(value) FROM counters`).Scan(&sum); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `UPDATE counters SET value = value + 1 WHERE name = 'b'`); err != nil {
			fnErr = err
			return err
		}
		if attempts == 1 {
			if err := h.commit(); err != nil {
				return fmt.Errorf("сосед: %w", err)
			}
		}
		return nil
	})
	if fnErr != nil {
		t.Fatalf("конфликт должен был всплыть на COMMIT, а UPDATE в fn упал: %v", fnErr)
	}
	if err != nil {
		t.Fatalf("WithRetry = %v (код %q), want nil: 40001 на COMMIT тоже надо повторять", err, code(err))
	}
	if attempts != 2 {
		t.Errorf("fn вызвана %d раз, want 2", attempts)
	}
	got := testdb.Rows(t, pool, `SELECT name, value FROM counters WHERE name IN ('a', 'b') ORDER BY name`)
	if want := [][]string{{"a", "1"}, {"b", "1"}}; !reflect.DeepEqual(got, want) {
		t.Errorf("counters = %v, want %v", got, want)
	}
}
