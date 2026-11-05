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
// Код ниже работает правильно, но делает 1 запрос за проектами и ещё по
// запросу на КАЖДЫЙ проект: классический N+1. На 20 проектах это 21 запрос,
// на 1000 проектов уже 1001 поход в базу.
//
// TODO: перепиши так, чтобы функция делала не больше 2 запросов к базе
// при любом числе проектов (тест считает запросы). Результат не должен измениться.
// Варианты: второй запрос WHERE project_id = ANY($1) и раскладка в Go
// (так делает Preload в GORM) или один запрос с LEFT JOIN.
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
	for i := range projects {
		rows, _ := pool.Query(ctx, `SELECT id, title, done FROM tasks WHERE project_id = $1 ORDER BY id`, projects[i].ID)
		projects[i].Tasks, err = pgx.CollectRows(rows, pgx.RowToStructByPos[Task])
		if err != nil {
			return nil, err
		}
	}
	return projects, nil
}
