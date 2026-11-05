package l08sqlcormmigrations

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"goschool/homework/internal/testdb"
)

func TestBuildTaskQuery(t *testing.T) {
	t.Parallel()
	pool := testdb.New(t, "testdata/seed.sql")
	id := func(v int64) *int64 { return &v }
	yes, no := true, false

	tests := []struct {
		name   string
		filter TaskFilter
		want   string // эталонный запрос с литералами, ответ сравниваем по id
	}{
		{"без фильтров", TaskFilter{}, `SELECT id FROM tasks ORDER BY id`},
		{"проект", TaskFilter{ProjectID: id(5)}, `SELECT id FROM tasks WHERE project_id = 5 ORDER BY id`},
		{"проект 0 это фильтр, а не «любой»", TaskFilter{ProjectID: id(0)}, `SELECT id FROM tasks WHERE false`},
		{"только незавершённые", TaskFilter{Done: &no}, `SELECT id FROM tasks WHERE NOT done ORDER BY id`},
		{"проект и завершённые", TaskFilter{ProjectID: id(1), Done: &yes}, `SELECT id FROM tasks WHERE project_id = 1 AND done ORDER BY id`},
		{"лимит", TaskFilter{ProjectID: id(3), Limit: 4}, `SELECT id FROM tasks WHERE project_id = 3 ORDER BY id LIMIT 4`},
		{"поиск без учёта регистра", TaskFilter{Search: "ЗАДАЧА 10"}, `SELECT id FROM tasks WHERE title ILIKE '%задача 10%' ORDER BY id`},
		{"кавычка в поиске", TaskFilter{Search: "o'brien"}, `SELECT 202`},
		{"процент буквально", TaskFilter{Search: "%"}, `SELECT 201`},
		{"инъекция", TaskFilter{Search: "' OR '1'='1"}, `SELECT id FROM tasks WHERE false`},
		{"все фильтры", TaskFilter{ProjectID: id(2), Done: &yes, Search: "задача", Limit: 3}, `SELECT id FROM tasks WHERE project_id = 2 AND done AND title ILIKE '%задача%' ORDER BY id LIMIT 3`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q, args := BuildTaskQuery(tt.filter)
			if q == "" {
				t.Fatalf("BuildTaskQuery(%s) вернула пустой запрос", desc(tt.filter))
			}
			if tt.filter.Search != "" && strings.Contains(q, tt.filter.Search) {
				t.Fatalf("значение Search попало в текст запроса: %s\nиспользуй плейсхолдер", q)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			rows, err := pool.Query(ctx, q, args...)
			if err != nil {
				t.Fatalf("запрос %q с args %v: %v", q, args, err)
			}
			type row struct {
				ID    int64
				Title string
				Done  bool
			}
			got, err := pgx.CollectRows(rows, pgx.RowToStructByPos[row])
			if err != nil {
				t.Fatalf("запрос %q с args %v: %v (колонки должны быть id, title, done)", q, args, err)
			}
			var gotIDs []string
			for _, r := range got {
				gotIDs = append(gotIDs, strconv.FormatInt(r.ID, 10))
			}
			var wantIDs []string
			for _, r := range testdb.Rows(t, pool, tt.want) {
				wantIDs = append(wantIDs, r[0])
			}
			if !reflect.DeepEqual(gotIDs, wantIDs) {
				t.Errorf("BuildTaskQuery(%s):\n%s %v\nвернул id %v, want %v", desc(tt.filter), q, args, gotIDs, wantIDs)
			}
		})
	}
}

// desc печатает фильтр с разыменованными указателями.
func desc(f TaskFilter) string {
	p, d := "nil", "nil"
	if f.ProjectID != nil {
		p = strconv.FormatInt(*f.ProjectID, 10)
	}
	if f.Done != nil {
		d = strconv.FormatBool(*f.Done)
	}
	return fmt.Sprintf("{ProjectID:%s Done:%s Search:%q Limit:%d}", p, d, f.Search, f.Limit)
}
