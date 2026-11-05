package l08sqlcormmigrations

import (
	"context"
	"reflect"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/homework/internal/testdb"
)

// queryCounter это трейсер pgx: pgx вызывает TraceQueryStart перед каждым запросом.
type queryCounter struct{ n atomic.Int64 }

func (c *queryCounter) TraceQueryStart(ctx context.Context, _ *pgx.Conn, _ pgx.TraceQueryStartData) context.Context {
	c.n.Add(1)
	return ctx
}

func (c *queryCounter) TraceQueryEnd(context.Context, *pgx.Conn, pgx.TraceQueryEndData) {}

func TestProjectsWithTasks(t *testing.T) {
	t.Parallel()
	seeded, url := testdb.NewWithURL(t, "testdata/seed.sql")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Ожидание строим одним JOIN-ом через отдельный пул.
	want := map[int64]*Project{}
	var order []int64
	for _, r := range testdb.Rows(t, seeded, `
		SELECT p.id, p.name, t.id, t.title, t.done
		FROM projects p LEFT JOIN tasks t ON t.project_id = p.id
		ORDER BY p.id, t.id`) {
		pid, _ := strconv.ParseInt(r[0], 10, 64)
		if want[pid] == nil {
			want[pid] = &Project{ID: pid, Name: r[1]}
			order = append(order, pid)
		}
		if r[2] != "NULL" {
			tid, _ := strconv.ParseInt(r[2], 10, 64)
			want[pid].Tasks = append(want[pid].Tasks, Task{ID: tid, Title: r[3], Done: r[4] == "true"})
		}
	}

	counter := &queryCounter{}
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.Tracer = counter
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	got, err := ProjectsWithTasks(ctx, pool)
	if err != nil {
		t.Fatalf("ProjectsWithTasks error = %v", err)
	}
	if len(got) != len(order) {
		t.Fatalf("ProjectsWithTasks вернула %d проектов, want %d", len(got), len(order))
	}
	for i, id := range order {
		w := want[id]
		g := got[i]
		if g.ID != w.ID || g.Name != w.Name || len(g.Tasks) != len(w.Tasks) ||
			(len(w.Tasks) > 0 && !reflect.DeepEqual(g.Tasks, w.Tasks)) {
			t.Fatalf("проект #%d = {%d %s, %d задач}, want {%d %s, %d задач}",
				i, g.ID, g.Name, len(g.Tasks), w.ID, w.Name, len(w.Tasks))
		}
	}
	if n := counter.n.Load(); n > 2 {
		t.Errorf("ProjectsWithTasks сделала %d запросов к базе, want не больше 2: это N+1", n)
	}
}
