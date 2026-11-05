package l01postgresintro

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"

	"goschool/homework/internal/testdb"
)

func TestQueries(t *testing.T) {
	runCases(t, []sqlCase{
		{
			file:    "q01_tables.sql",
			ordered: true,
			want: [][]string{
				{"authors"}, {"books"}, {"categories"}, {"customers"}, {"order_items"}, {"orders"},
			},
			hint: "без фильтра table_schema = current_schema() в результат попадут системные таблицы; нужен ORDER BY table_name",
		},
		{
			file:    "q02_columns.sql",
			ordered: true,
			want: [][]string{
				{"id", "bigint", "NO"},
				{"title", "text", "NO"},
				{"author_id", "bigint", "YES"},
				{"category_id", "bigint", "YES"},
				{"price", "numeric", "YES"},
				{"year", "integer", "YES"},
				{"pages", "integer", "NO"},
			},
			hint: "WHERE table_schema = current_schema() AND table_name = 'books' ORDER BY ordinal_position",
		},
	})
}

func TestCreateTable(t *testing.T) {
	pool := testdb.New(t, seed)
	testdb.Exec(t, pool, testdb.ReadQuery(t, "q03_create_table.sql"))

	expectRows(t, pool, "колонки wishlist не совпадают с условием (имя, тип, NULL, identity)",
		`SELECT column_name, data_type, is_nullable, is_identity
		 FROM information_schema.columns
		 WHERE table_schema = current_schema() AND table_name = 'wishlist'
		 ORDER BY ordinal_position`,
		[][]string{
			{"id", "bigint", "NO", "YES"},
			{"customer_id", "bigint", "NO", "NO"},
			{"title", "text", "NO", "NO"},
			{"added_at", "timestamp with time zone", "NO", "NO"},
			{"note", "text", "YES", "NO"},
		})
	expectRows(t, pool, "id должен быть первичным ключом",
		`SELECT count(*) FROM information_schema.table_constraints
		 WHERE table_schema = current_schema() AND table_name = 'wishlist' AND constraint_type = 'PRIMARY KEY'`,
		[][]string{{"1"}})
	expectRows(t, pool, "вставка без id и added_at: id выдаёт база, added_at заполняется now()",
		`INSERT INTO wishlist (customer_id, title) VALUES (1, 'Дюна') RETURNING id, added_at IS NOT NULL, note IS NULL`,
		[][]string{{"1", "true", "true"}})

	_, err := pool.Exec(context.Background(), `INSERT INTO wishlist (customer_id, title) VALUES (1, NULL)`)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23502" {
		t.Fatalf("вставка title = NULL: ошибка %v, want not_null_violation (23502)", err)
	}
}
