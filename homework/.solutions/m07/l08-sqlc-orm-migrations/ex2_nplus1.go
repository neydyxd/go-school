package l08sqlcormmigrations

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Task struct {
	ID    int64
	Title string
	Done  bool
}

type Project struct {
	ID    int64
	Name  string
	Tasks []Task
}

// ProjectsWithTasks возвращает все проекты (по возрастанию id), у каждого
// задачи по возрастанию id. У проекта без задач Tasks пустой.
//
// Два запроса при любом числе проектов: проекты, затем все их задачи
// одним WHERE project_id = ANY($1). Так работает Preload в GORM.
func ProjectsWithTasks(ctx context.Context, pool *pgxpool.Pool) ([]Project, error) {
	rows, _ := pool.Query(ctx, `SELECT id, name FROM projects ORDER BY id`)
	projects, err := pgx.CollectRows(rows, func(r pgx.CollectableRow) (Project, error) {
		var p Project
		err := r.Scan(&p.ID, &p.Name)
		return p, err
	})
	if err != nil {
		return nil, err
	}

	ids := make([]int64, len(projects))
	byID := make(map[int64]*Project, len(projects))
	for i := range projects {
		ids[i] = projects[i].ID
		byID[projects[i].ID] = &projects[i]
	}

	// pgx сам превращает []int64 в массив bigint[] для ANY.
	rows, _ = pool.Query(ctx, `SELECT project_id, id, title, done FROM tasks WHERE project_id = ANY($1) ORDER BY id`, ids)
	var projectID int64
	var t Task
	_, err = pgx.ForEachRow(rows, []any{&projectID, &t.ID, &t.Title, &t.Done}, func() error {
		p := byID[projectID]
		p.Tasks = append(p.Tasks, t)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return projects, nil
}
