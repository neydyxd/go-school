// Package testdb даёт каждому тесту свою изолированную схему в учебном PostgreSQL.
//
// База поднимается из папки homework: docker compose up -d.
// Адрес можно переопределить переменной GOSCHOOL_PG_URL.
package testdb

import (
	"context"
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultURL = "postgres://goschool:goschool@localhost:55432/goschool?sslmode=disable"

const hint = "подними базу: cd homework && docker compose up -d"

// BaseURL возвращает адрес учебной базы.
func BaseURL() string {
	if u := os.Getenv("GOSCHOOL_PG_URL"); u != "" {
		return u
	}
	return defaultURL
}

// New создаёт схему только для этого теста, выполняет seed-файлы и возвращает пул,
// у которого search_path указывает на эту схему. Схема удаляется после теста.
func New(t testing.TB, seedFiles ...string) *pgxpool.Pool {
	t.Helper()
	pool, _ := NewWithURL(t, seedFiles...)
	return pool
}

// NewWithURL как New, но дополнительно возвращает строку подключения с той же схемой
// (для database/sql, goose и других инструментов).
func NewWithURL(t testing.TB, seedFiles ...string) (*pgxpool.Pool, string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	admin, err := pgx.Connect(ctx, BaseURL())
	if err != nil {
		t.Fatalf("PostgreSQL недоступен (%v): %s", err, hint)
	}
	defer admin.Close(context.Background())

	schema := "t_" + randHex(6)
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+schema); err != nil {
		t.Fatalf("не удалось создать схему: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		c, err := pgx.Connect(ctx, BaseURL())
		if err != nil {
			return
		}
		defer c.Close(ctx)
		_, _ = c.Exec(ctx, "DROP SCHEMA IF EXISTS "+schema+" CASCADE")
	})

	connURL := withSearchPath(t, BaseURL(), schema)
	cfg, err := pgxpool.ParseConfig(connURL)
	if err != nil {
		t.Fatalf("плохой адрес базы: %v", err)
	}
	cfg.MaxConns = 4 // ponytail: тестов много и они параллельны, на пул хватает четырёх соединений
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		t.Fatalf("PostgreSQL недоступен (%v): %s", err, hint)
	}
	t.Cleanup(pool.Close) // регистрируется после DROP SCHEMA, значит выполнится раньше него

	for _, f := range seedFiles {
		Exec(t, pool, ReadFile(t, f))
	}
	return pool, connURL
}

// Exec выполняет один или несколько SQL-операторов через простой протокол.
func Exec(t testing.TB, pool *pgxpool.Pool, sql string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if _, err := pool.Exec(ctx, sql, pgx.QueryExecModeSimpleProtocol); err != nil {
		t.Fatalf("ошибка выполнения SQL: %v\n%s", err, sql)
	}
}

// ReadFile читает файл целиком или валит тест.
func ReadFile(t testing.TB, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("не удалось прочитать %s: %v", path, err)
	}
	return string(b)
}

// ReadQuery читает файл с запросом студента. Если в нём только комментарии
// или пусто, тест падает с понятным сообщением.
func ReadQuery(t testing.TB, path string) string {
	t.Helper()
	q := ReadFile(t, path)
	if strings.TrimSpace(stripComments(q)) == "" {
		t.Fatalf("%s: запрос ещё не написан (в файле только комментарии)", path)
	}
	return q
}

// Rows выполняет запрос и возвращает строки результата, где каждое значение
// приведено к строке (NULL становится "NULL"). Порядок строк как в результате.
func Rows(t testing.TB, pool *pgxpool.Pool, sql string, args ...any) [][]string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		t.Fatalf("ошибка запроса: %v\n%s", err, sql)
	}
	defer rows.Close()
	var out [][]string
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			t.Fatalf("ошибка чтения строки: %v", err)
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			row[i] = format(v)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("ошибка запроса: %v\n%s", err, sql)
	}
	return out
}

func format(v any) string {
	switch x := v.(type) {
	case nil:
		return "NULL"
	case time.Time:
		return x.UTC().Format(time.RFC3339)
	case [16]byte: // uuid
		return fmt.Sprintf("%x-%x-%x-%x-%x", x[0:4], x[4:6], x[6:8], x[8:10], x[10:16])
	case driver.Valuer: // pgtype.Numeric и другие типы pgx отдают текстовое значение
		dv, err := x.Value()
		if err != nil {
			return fmt.Sprint(x)
		}
		return format(dv)
	case fmt.Stringer:
		return x.String()
	default:
		return fmt.Sprint(x)
	}
}

func stripComments(q string) string {
	var b strings.Builder
	for _, line := range strings.Split(q, "\n") {
		if i := strings.Index(line, "--"); i >= 0 {
			line = line[:i]
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String()
}

func withSearchPath(t testing.TB, base, schema string) string {
	t.Helper()
	u, err := url.Parse(base)
	if err != nil {
		t.Fatalf("плохой GOSCHOOL_PG_URL: %v", err)
	}
	q := u.Query()
	q.Set("search_path", schema)
	u.RawQuery = q.Encode()
	return u.String()
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
