package l08sqlcormmigrations

import (
	"fmt"
	"strings"
)

// TaskFilter это фильтры списка задач из query-параметров API.
// nil и пустые значения означают «не фильтровать».
type TaskFilter struct {
	ProjectID *int64 // указатель: ProjectID = 0 это фильтр «проект 0», а nil это «любой»
	Done      *bool  // nil: любые, &false: только незавершённые
	Search    string // подстрока в title без учёта регистра
	Limit     int    // 0: без лимита
}

// BuildTaskQuery собирает запрос
//
//	SELECT id, title, done FROM tasks [WHERE ... AND ...] ORDER BY id [LIMIT $N]
//
// и аргументы к нему. Так под капотом работают query builder'ы вроде squirrel.
//
//   - каждое значение идёт в args, а в текст запроса попадает только
//     плейсхолдер $1, $2, ... по порядку. Значения в текст не вставлять:
//     тест проверит, что в SQL нет кусков Search;
//   - Search ищется буквально (символы % и _ ничего не значат), без учёта
//     регистра. Проще всего strpos(lower(title), lower($N)) > 0;
//   - если ни одного фильтра нет, WHERE не пишется вовсе.
func BuildTaskQuery(f TaskFilter) (string, []any) {
	var where []string
	var args []any
	// arg кладёт значение в args и возвращает его плейсхолдер.
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}
	if f.ProjectID != nil {
		where = append(where, "project_id = "+arg(*f.ProjectID))
	}
	if f.Done != nil {
		where = append(where, "done = "+arg(*f.Done))
	}
	if f.Search != "" {
		where = append(where, "strpos(lower(title), lower("+arg(f.Search)+")) > 0")
	}

	var b strings.Builder
	b.WriteString("SELECT id, title, done FROM tasks")
	if len(where) > 0 {
		b.WriteString(" WHERE " + strings.Join(where, " AND "))
	}
	b.WriteString(" ORDER BY id")
	if f.Limit > 0 {
		b.WriteString(" LIMIT " + arg(f.Limit))
	}
	return b.String(), args
}
