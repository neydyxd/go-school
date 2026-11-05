package l05indexes

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/homework/internal/testdb"
)

// planNode это узел плана из EXPLAIN (FORMAT JSON).
type planNode struct {
	NodeType  string     `json:"Node Type"`
	IndexName string     `json:"Index Name"`
	Plans     []planNode `json:"Plans"`
}

// explain возвращает все узлы плана списком (обход в глубину). settings это
// параметры планировщика (например "enable_seqscan = off"): они ставятся через
// SET LOCAL в отдельной транзакции и не влияют на другие запросы.
func explain(t *testing.T, pool *pgxpool.Pool, query string, settings ...string) []planNode {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("BEGIN: %v", err)
	}
	defer tx.Rollback(ctx)
	for _, s := range settings {
		if _, err := tx.Exec(ctx, "SET LOCAL "+s); err != nil {
			t.Fatalf("SET LOCAL %s: %v", s, err)
		}
	}
	var raw string
	if err := tx.QueryRow(ctx, "EXPLAIN (FORMAT JSON) "+query).Scan(&raw); err != nil {
		t.Fatalf("EXPLAIN упал: %v\n%s", err, query)
	}
	var doc []struct {
		Plan planNode `json:"Plan"`
	}
	if err := json.Unmarshal([]byte(raw), &doc); err != nil || len(doc) == 0 {
		t.Fatalf("не разобрал EXPLAIN: %v", err)
	}
	var out []planNode
	var walk func(n planNode)
	walk = func(n planNode) {
		out = append(out, n)
		for _, c := range n.Plans {
			walk(c)
		}
	}
	walk(doc[0].Plan)
	return out
}

// planText нужен для понятного сообщения об ошибке.
func planText(t *testing.T, pool *pgxpool.Pool, query string) string {
	t.Helper()
	var lines []string
	for _, r := range testdb.Rows(t, pool, "EXPLAIN (COSTS OFF) "+query) {
		lines = append(lines, r[0])
	}
	return strings.Join(lines, "\n")
}

func nodeTypes(nodes []planNode) []string {
	var out []string
	for _, n := range nodes {
		out = append(out, n.NodeType)
	}
	return out
}

func hasNode(nodes []planNode, types ...string) bool {
	for _, n := range nodes {
		if slices.Contains(types, n.NodeType) {
			return true
		}
	}
	return false
}

// usedIndexes возвращает индексы из плана, кроме первичного ключа.
func usedIndexes(nodes []planNode) []string {
	var out []string
	for _, n := range nodes {
		if n.IndexName != "" && n.IndexName != "orders_pkey" {
			out = append(out, n.IndexName)
		}
	}
	return out
}

type indexInfo struct {
	method  string // btree, gin...
	partial bool
}

func describeIndex(t *testing.T, pool *pgxpool.Pool, name string) indexInfo {
	t.Helper()
	r := testdb.Rows(t, pool, `
		SELECT am.amname, (i.indpred IS NOT NULL)::text
		FROM pg_index i
		JOIN pg_class c ON c.oid = i.indexrelid
		JOIN pg_am am ON am.oid = c.relam
		WHERE c.oid = to_regclass($1)`, name)
	if len(r) != 1 {
		t.Fatalf("индекс %s не найден в каталоге", name)
	}
	return indexInfo{method: r[0][0], partial: r[0][1] == "true"}
}

// withStudentIndex накатывает сид, выполняет файл студента и обновляет статистику.
// Возвращает пул и результат запроса до и после файла студента.
func withStudentIndex(t *testing.T, file, query string) (pool *pgxpool.Pool, before, after [][]string) {
	t.Helper()
	pool = testdb.New(t, "testdata/seed.sql")
	before = testdb.Rows(t, pool, query)
	testdb.Exec(t, pool, testdb.ReadQuery(t, file))
	testdb.Exec(t, pool, "ANALYZE orders")
	after = testdb.Rows(t, pool, query)
	return pool, before, after
}

// keyColumns возвращает ключевые колонки индекса по порядку (без INCLUDE).
func keyColumns(t *testing.T, pool *pgxpool.Pool, name string) []string {
	t.Helper()
	var out []string
	for _, r := range testdb.Rows(t, pool, `
		SELECT pg_get_indexdef(i.indexrelid, k, true)
		FROM pg_index i, generate_series(1, i.indnkeyatts) AS k
		WHERE i.indexrelid = to_regclass($1)
		ORDER BY k`, name) {
		out = append(out, r[0])
	}
	return out
}

// leadingIs проверяет, что все использованные индексы начинаются с колонок cols.
func leadingIs(t *testing.T, pool *pgxpool.Pool, used []string, cols ...string) string {
	t.Helper()
	for _, ix := range used {
		keys := keyColumns(t, pool, ix)
		if len(keys) < len(cols) || !slices.Equal(keys[:len(cols)], cols) {
			return fmt.Sprintf("у индекса %s ключевые колонки %v, а должны начинаться с %v: "+
				"с другой ведущей колонкой Postgres читает индекс почти целиком", ix, keys, cols)
		}
	}
	return ""
}

func sorted(rows [][]string) [][]string {
	out := slices.Clone(rows)
	slices.SortFunc(out, func(a, b []string) int { return slices.Compare(a, b) })
	return out
}

func TestIndexes(t *testing.T) {
	tests := []struct {
		name  string
		file  string
		query string
		check func(t *testing.T, pool *pgxpool.Pool, nodes []planNode) string // "" если всё хорошо
	}{
		{
			name:  "составной индекс убирает сортировку",
			file:  "q01_customer_recent.sql",
			query: `SELECT id, created_at, amount FROM orders WHERE customer_id = 42 ORDER BY created_at DESC LIMIT 10`,
			check: func(t *testing.T, pool *pgxpool.Pool, nodes []planNode) string {
				if hasNode(nodes, "Seq Scan", "Sort", "Incremental Sort") {
					return "в плане осталась сортировка или Seq Scan: нужен индекс, где после customer_id идёт created_at"
				}
				used := usedIndexes(nodes)
				if len(used) == 0 {
					return "план не использует твой индекс"
				}
				return leadingIs(t, pool, used, "customer_id", "created_at")
			},
		},
		{
			name:  "покрывающий индекс даёт Index Only Scan",
			file:  "q02_covering.sql",
			query: `SELECT customer_id, sum(amount) FROM orders WHERE customer_id BETWEEN 100 AND 199 GROUP BY customer_id`,
			check: func(t *testing.T, pool *pgxpool.Pool, nodes []planNode) string {
				// Index Only Scan в обычном плане зависит от карты видимости, а её
				// обновляет VACUUM, которому мешает любая долгая транзакция в базе.
				// Поэтому запрещаем Seq Scan и Bitmap: если индекс покрывающий,
				// планировщик выберет Index Only Scan при любой карте видимости.
				nodes = explain(t, pool, `SELECT customer_id, sum(amount) FROM orders WHERE customer_id BETWEEN 100 AND 199 GROUP BY customer_id`,
					"enable_seqscan = off", "enable_bitmapscan = off")
				if !hasNode(nodes, "Index Only Scan") {
					return "нет Index Only Scan: в индексе должны быть все колонки запроса (customer_id и amount)"
				}
				if msg := leadingIs(t, pool, usedIndexes(nodes), "customer_id"); msg != "" {
					return msg
				}
				return ""
			},
		},
		{
			name:  "частичный индекс для pending",
			file:  "q03_partial.sql",
			query: `SELECT id, customer_id, created_at FROM orders WHERE status = 'pending' ORDER BY created_at LIMIT 20`,
			check: func(t *testing.T, pool *pgxpool.Pool, nodes []planNode) string {
				if hasNode(nodes, "Seq Scan", "Sort", "Incremental Sort") {
					return "в плане осталась сортировка или Seq Scan: индекс должен отдавать строки в порядке created_at"
				}
				used := usedIndexes(nodes)
				if len(used) == 0 {
					return "план не использует твой индекс"
				}
				for _, ix := range used {
					if !describeIndex(t, pool, ix).partial {
						return "индекс " + ix + " построен по всем строкам, а нужен частичный (CREATE INDEX ... WHERE status = 'pending')"
					}
				}
				return ""
			},
		},
		{
			name:  "индекс по выражению lower(email)",
			file:  "q04_lower_email.sql",
			query: `SELECT id FROM orders WHERE lower(email) = 'user42@mail.example' ORDER BY id`,
			check: func(t *testing.T, pool *pgxpool.Pool, nodes []planNode) string {
				if hasNode(nodes, "Seq Scan") || len(usedIndexes(nodes)) == 0 {
					return "запрос читает всю таблицу: индекс должен быть построен по выражению lower(email)"
				}
				return ""
			},
		},
		{
			name:  "GIN для jsonb",
			file:  "q05_jsonb_gin.sql",
			query: `SELECT count(*) FROM orders WHERE attrs @> '{"promo": "BLACKFRIDAY"}'`,
			check: func(t *testing.T, pool *pgxpool.Pool, nodes []planNode) string {
				used := usedIndexes(nodes)
				if hasNode(nodes, "Seq Scan") || len(used) == 0 {
					return "запрос читает всю таблицу: нужен индекс USING gin (attrs)"
				}
				for _, ix := range used {
					if m := describeIndex(t, pool, ix).method; m != "gin" {
						return "индекс " + ix + " типа " + m + ", а нужен gin"
					}
				}
				return ""
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool, before, after := withStudentIndex(t, tt.file, tt.query)
			// порядок строк без ORDER BY не задан, сравниваем как множества
			if !reflect.DeepEqual(sorted(before), sorted(after)) {
				t.Fatalf("после %s запрос вернул другие строки: было %v, стало %v", tt.file, before, after)
			}
			nodes := explain(t, pool, tt.query)
			if msg := tt.check(t, pool, nodes); msg != "" {
				t.Errorf("%s: %s\nузлы плана: %v\nплан:\n%s", tt.file, msg, nodeTypes(nodes), planText(t, pool, tt.query))
			}
		})
	}
}

// rowsInTZ выполняет запрос в транзакции с другим часовым поясом сессии:
// границы дня без явного пояса (+00) тогда сдвинутся, и ответ изменится.
func rowsInTZ(t *testing.T, pool *pgxpool.Pool, tz, query string) [][]string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("BEGIN: %v", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SET LOCAL TimeZone = '"+tz+"'"); err != nil {
		t.Fatal(err)
	}
	// простой протокол: без кеша подготовленных запросов, иначе литерал даты,
	// разобранный при первой подготовке, не перечитается в новом поясе
	rows, err := tx.Query(ctx, query, pgx.QueryExecModeSimpleProtocol)
	if err != nil {
		t.Fatalf("ошибка запроса: %v\n%s", err, query)
	}
	defer rows.Close()
	var out [][]string
	for rows.Next() {
		vals, err := rows.Values()
		if err != nil {
			t.Fatal(err)
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			row[i] = fmt.Sprint(v)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("ошибка запроса: %v\n%s", err, query)
	}
	return out
}

func TestSargable(t *testing.T) {
	t.Parallel()
	pool := testdb.New(t, "testdata/seed.sql")
	testdb.Exec(t, pool, "CREATE INDEX orders_created_idx ON orders (created_at)")
	testdb.Exec(t, pool, "ANALYZE orders")

	query := testdb.ReadQuery(t, "q06_sargable.sql")
	want := rowsInTZ(t, pool, "UTC", `SELECT count(*), min(id), max(id) FROM orders WHERE (created_at AT TIME ZONE 'UTC')::date = '2026-03-15'`)
	for _, tz := range []string{"UTC", "Europe/Moscow", "America/New_York"} {
		if got := rowsInTZ(t, pool, tz, query); !reflect.DeepEqual(got, want) {
			t.Fatalf("q06_sargable.sql при TimeZone = %s вернул %v, want %v: границы дня задай с явным поясом UTC (+00)", tz, got, want)
		}
	}
	nodes := explain(t, pool, query)
	if hasNode(nodes, "Seq Scan") || !slices.Contains(usedIndexes(nodes), "orders_created_idx") {
		t.Errorf("q06_sargable.sql не использует orders_created_idx: сравнивай саму колонку created_at с границами дня\nузлы плана: %v\nплан:\n%s",
			nodeTypes(nodes), planText(t, pool, query))
	}
	if n := testdb.Rows(t, pool, `SELECT count(*) FROM pg_indexes WHERE schemaname = current_schema() AND tablename = 'orders'`); n[0][0] != "2" {
		t.Errorf("в схеме %s индексов, а должно быть 2 (pkey и orders_created_idx): новые индексы не создавай", n[0][0])
	}
}
