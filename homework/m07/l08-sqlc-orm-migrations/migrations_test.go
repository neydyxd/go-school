package l08sqlcormmigrations

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib" // драйвер "pgx" для database/sql, на нём работает goose
	"github.com/pressly/goose/v3"

	"goschool/homework/internal/testdb"
)

func columns(t *testing.T, db *sql.DB, table string) [][]string {
	t.Helper()
	rows, err := db.Query(`
		SELECT column_name, data_type, is_nullable, is_identity
		FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = $1
		ORDER BY ordinal_position`, table)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out [][]string
	for rows.Next() {
		var a, b, c, d string
		if err := rows.Scan(&a, &b, &c, &d); err != nil {
			t.Fatal(err)
		}
		out = append(out, []string{a, b, c, d})
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

func mustExec(t *testing.T, db *sql.DB, q string, args ...any) {
	t.Helper()
	if _, err := db.Exec(q, args...); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
}

func execCode(db *sql.DB, q string) string {
	_, err := db.Exec(q)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	if err != nil {
		return err.Error()
	}
	return "нет ошибки"
}

func queryString(t *testing.T, db *sql.DB, q string) string {
	t.Helper()
	var s string
	if err := db.QueryRow(q).Scan(&s); err != nil {
		t.Fatalf("%s: %v", q, err)
	}
	return s
}

var (
	wantProjects = [][]string{
		{"id", "bigint", "NO", "YES"},
		{"name", "text", "NO", "NO"},
		{"created_at", "timestamp with time zone", "NO", "NO"},
	}
	wantTasksV2 = [][]string{
		{"id", "bigint", "NO", "YES"},
		{"project_id", "bigint", "NO", "NO"},
		{"title", "text", "NO", "NO"},
		{"done", "boolean", "NO", "NO"},
	}
	wantTasksV3 = append(append([][]string{}, wantTasksV2...), []string{"priority", "smallint", "NO", "NO"})
)

func TestMigrations(t *testing.T) {
	t.Parallel()
	_, url := testdb.NewWithURL(t)
	db, err := sql.Open("pgx", url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	p, err := goose.NewProvider(goose.DialectPostgres, db, os.DirFS("migrations"))
	if err != nil {
		t.Fatalf("goose.NewProvider: %v", err)
	}

	// 1. Накатываем первые две миграции и проверяем схему.
	if _, err := p.UpTo(ctx, 2); err != nil {
		t.Fatalf("up до версии 2: %v", err)
	}
	if got := columns(t, db, "projects"); !reflect.DeepEqual(got, wantProjects) {
		t.Fatalf("колонки projects = %v, want %v", got, wantProjects)
	}
	if got := columns(t, db, "tasks"); !reflect.DeepEqual(got, wantTasksV2) {
		t.Fatalf("колонки tasks после миграции 2 = %v, want %v", got, wantTasksV2)
	}
	if n := queryString(t, db, `SELECT count(*)::text FROM pg_indexes WHERE schemaname = current_schema() AND indexname = 'tasks_project_id_idx'`); n != "1" {
		t.Errorf("нет индекса tasks_project_id_idx")
	}

	// 2. Поведение: DEFAULT, UNIQUE, внешний ключ с каскадом.
	mustExec(t, db, `INSERT INTO projects (name) VALUES ('Альфа'), ('Бета')`)
	mustExec(t, db, `INSERT INTO tasks (project_id, title) VALUES (1, 'a1'), (1, 'a2'), (2, 'b1')`)
	if s := queryString(t, db, `SELECT (created_at IS NOT NULL AND NOT (SELECT bool_or(done) FROM tasks))::text FROM projects WHERE id = 1`); s != "true" {
		t.Errorf("DEFAULT не сработал: created_at должен заполняться now(), done по умолчанию false")
	}
	if code := execCode(db, `INSERT INTO projects (name) VALUES ('Альфа')`); code != "23505" {
		t.Errorf("повторное имя проекта: %s, want ошибку 23505 (UNIQUE)", code)
	}
	if code := execCode(db, `INSERT INTO tasks (project_id, title) VALUES (99, 'x')`); code != "23503" {
		t.Errorf("задача несуществующего проекта: %s, want ошибку 23503 (внешний ключ)", code)
	}
	mustExec(t, db, `DELETE FROM projects WHERE id = 2`)
	if n := queryString(t, db, `SELECT count(*)::text FROM tasks`); n != "2" {
		t.Errorf("после удаления проекта 2 задач %s, want 2: нужен ON DELETE CASCADE", n)
	}

	// 3. Третья миграция на таблице с данными.
	if _, err := p.Up(ctx); err != nil {
		t.Fatalf("up до последней версии: %v", err)
	}
	if got := columns(t, db, "tasks"); !reflect.DeepEqual(got, wantTasksV3) {
		t.Fatalf("колонки tasks после миграции 3 = %v, want %v", got, wantTasksV3)
	}
	if s := queryString(t, db, `SELECT string_agg(priority::text, ',' ORDER BY id) FROM tasks`); s != "0,0" {
		t.Errorf("priority у старых задач = %s, want 0,0", s)
	}
	if code := execCode(db, `INSERT INTO tasks (project_id, title, priority) VALUES (1, 'x', 4)`); code != "23514" {
		t.Errorf("priority = 4: %s, want ошибку 23514 (CHECK)", code)
	}
	if n := queryString(t, db, `SELECT count(*)::text FROM pg_constraint WHERE conname = 'tasks_priority_check' AND connamespace = current_schema()::regnamespace`); n != "1" {
		t.Errorf("нет ограничения tasks_priority_check")
	}

	// 4. Down по одной: каждая миграция отменяется полностью.
	if _, err := p.Down(ctx); err != nil {
		t.Fatalf("down миграции 3: %v", err)
	}
	if got := columns(t, db, "tasks"); !reflect.DeepEqual(got, wantTasksV2) {
		t.Errorf("колонки tasks после down 3 = %v, want %v", got, wantTasksV2)
	}
	if _, err := p.DownTo(ctx, 0); err != nil {
		t.Fatalf("down до нуля: %v", err)
	}
	if n := queryString(t, db, `SELECT count(*)::text FROM information_schema.tables WHERE table_schema = current_schema() AND table_name IN ('projects', 'tasks')`); n != "0" {
		t.Errorf("после down до нуля осталось таблиц: %s, want 0", n)
	}

	// 5. И снова вверх: down не оставил мусора, который мешает повторному up.
	if _, err := p.Up(ctx); err != nil {
		t.Fatalf("повторный up: %v", err)
	}
	if v, err := p.GetDBVersion(ctx); err != nil || v != 3 {
		t.Errorf("версия схемы после повторного up = %d (%v), want 3", v, err)
	}
}
