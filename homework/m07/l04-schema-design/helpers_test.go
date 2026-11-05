package l04schemadesign

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/homework/internal/testdb"
)

// Коды ошибок PostgreSQL (SQLSTATE), которые проверяют тесты.
const (
	notNullViolation    = "23502"
	foreignKeyViolation = "23503"
	uniqueViolation     = "23505"
	checkViolation      = "23514"
	invalidTextRepr     = "22P02" // неверное значение для enum-типа
	generatedAlways     = "428C9" // явный id в колонке GENERATED ALWAYS AS IDENTITY
)

// setup применяет schema.sql студента к пустой схеме.
func setup(t *testing.T) *pgxpool.Pool {
	t.Helper()
	t.Parallel()
	pool := testdb.New(t)
	testdb.Exec(t, pool, testdb.ReadQuery(t, "schema.sql"))
	return pool
}

// ok выполняет оператор, который схема обязана пропустить.
func ok(t *testing.T, pool *pgxpool.Pool, sql string) {
	t.Helper()
	if _, err := exec(pool, sql); err != nil {
		t.Fatalf("база отвергла корректный оператор: %v\n%s", err, sql)
	}
}

// fails выполняет оператор, который схема обязана отвергнуть с одним из кодов.
func fails(t *testing.T, pool *pgxpool.Pool, why, sql string, codes ...string) {
	t.Helper()
	_, err := exec(pool, sql)
	var pgErr *pgconn.PgError
	switch {
	case err == nil:
		t.Fatalf("%s: база приняла оператор, а должна была отвергнуть (ждали код %s)\n%s",
			why, strings.Join(codes, " или "), sql)
	case !errors.As(err, &pgErr):
		t.Fatalf("%s: не ошибка PostgreSQL: %v", why, err)
	case !slices.Contains(codes, pgErr.Code):
		t.Fatalf("%s: код ошибки %s (%s), want %s\n%s",
			why, pgErr.Code, pgErr.Message, strings.Join(codes, " или "), sql)
	}
}

func exec(pool *pgxpool.Pool, sql string) (pgconn.CommandTag, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return pool.Exec(ctx, sql)
}

// one возвращает единственное значение запроса.
func one(t *testing.T, pool *pgxpool.Pool, sql string) string {
	t.Helper()
	rows := testdb.Rows(t, pool, sql)
	if len(rows) != 1 || len(rows[0]) != 1 {
		t.Fatalf("ожидалось одно значение, получено %v\n%s", rows, sql)
	}
	return rows[0][0]
}

// expect сравнивает единственное значение запроса с ожидаемым.
func expect(t *testing.T, pool *pgxpool.Pool, why, sql, want string) {
	t.Helper()
	if got := one(t, pool, sql); got != want {
		t.Fatalf("%s: %s = %s, want %s", why, sql, got, want)
	}
}

// columnType возвращает тип колонки так, как его видит information_schema.
func columnType(t *testing.T, pool *pgxpool.Pool, table, column string) string {
	t.Helper()
	rows := testdb.Rows(t, pool, `SELECT data_type FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2`, table, column)
	if len(rows) == 0 {
		t.Fatalf("нет колонки %s.%s", table, column)
	}
	return rows[0][0]
}

// checkID проверяет, что id таблицы это bigint с identity, как требует спецификация.
func checkID(t *testing.T, pool *pgxpool.Pool, table string) {
	t.Helper()
	got := testdb.Rows(t, pool, `SELECT data_type, is_identity, coalesce(identity_generation, '-') FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = $1 AND column_name = 'id'`, table)
	if len(got) != 1 || got[0][0] != "bigint" || got[0][1] != "YES" || got[0][2] != "ALWAYS" {
		t.Fatalf("%s.id: %v, want bigint GENERATED ALWAYS AS IDENTITY (не serial, не int, не BY DEFAULT)", table, got)
	}
}
