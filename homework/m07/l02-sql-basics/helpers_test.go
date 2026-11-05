package l02sqlbasics

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/homework/internal/testdb"
)

const seed = "testdata/seed.sql"

// sqlCase описывает одну SQL-задачу: файл студента и ожидаемый результат.
type sqlCase struct {
	file    string
	ordered bool       // условие требует порядок: сравниваем строки с учётом порядка
	want    [][]string // ожидаемые строки, NULL записан как "NULL"
	hint    string     // типичная ошибка, если результат не совпал
	// before выполняется после сида и до запроса: добавляет строки-подвохи только для этой задачи
	before string
	// after выполняется после запроса: проверяет состояние таблиц (для INSERT/UPDATE/DELETE)
	after func(t *testing.T, pool *pgxpool.Pool)
}

// runCases запускает каждую задачу в своей схеме со свежим сидом.
func runCases(t *testing.T, cases []sqlCase) {
	for _, c := range cases {
		t.Run(strings.TrimSuffix(c.file, ".sql"), func(t *testing.T) {
			t.Parallel()
			pool := testdb.New(t, seed)
			if c.before != "" {
				testdb.Exec(t, pool, c.before)
			}
			got := testdb.Rows(t, pool, testdb.ReadQuery(t, c.file))
			if !sameRows(got, c.want, c.ordered) {
				extra := ""
				if c.before != "" {
					extra = "\nк сиду тест добавил: " + c.before
				}
				t.Fatalf("%s вернул:\n%s\nа ожидалось:\n%s\nподсказка: %s%s",
					c.file, show(got), show(c.want), c.hint, extra)
			}
			if c.after != nil {
				c.after(t, pool)
			}
		})
	}
}

// expectRows выполняет проверочный запрос и сравнивает результат с учётом порядка.
func expectRows(t *testing.T, pool *pgxpool.Pool, why, sql string, want [][]string) {
	t.Helper()
	if got := testdb.Rows(t, pool, sql); !sameRows(got, want, true) {
		t.Fatalf("%s\nзапрос: %s\nвернул:\n%s\nа ожидалось:\n%s", why, sql, show(got), show(want))
	}
}

func sameRows(got, want [][]string, ordered bool) bool {
	if len(got) != len(want) {
		return false
	}
	g, w := keys(got), keys(want)
	if !ordered {
		slices.Sort(g)
		slices.Sort(w)
	}
	return slices.Equal(g, w)
}

func keys(rows [][]string) []string {
	out := make([]string, len(rows))
	for i, r := range rows {
		out[i] = strings.Join(r, " | ")
	}
	return out
}

func show(rows [][]string) string {
	if len(rows) == 0 {
		return "  (пусто)"
	}
	var b strings.Builder
	for _, k := range keys(rows) {
		fmt.Fprintf(&b, "  %s\n", k)
	}
	return strings.TrimRight(b.String(), "\n")
}
