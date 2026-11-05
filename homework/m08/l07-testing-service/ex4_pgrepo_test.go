package l07testingservice

import (
	"context"
	"strings"
	"testing"
	"time"

	"goschool/homework/internal/testdb"
)

// Каждый подтест получает свою схему PostgreSQL: тесты не видят данных друг друга
// и могут идти параллельно. Нужна поднятая база: cd homework && docker compose up -d
func newPgRepo(t *testing.T) TaskRepo {
	return NewPgRepo(testdb.New(t, "testdata/schema.sql"))
}

func TestPgRepoContract(t *testing.T) {
	testRepoContract(t, newPgRepo)
}

func TestPgRepoDatabaseSpecific(t *testing.T) {
	ctx := context.Background()
	pool := testdb.New(t, "testdata/schema.sql")
	r := NewPgRepo(pool)

	before := time.Now().Add(-time.Minute)
	task, err := r.Create(ctx, Task{OwnerID: 1, Title: "x"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if task.CreatedAt.Before(before) {
		t.Errorf("CreatedAt = %v: want время из базы (DEFAULT now()), а не нулевое", task.CreatedAt)
	}

	t.Run("CHECK-ограничение статуса доходит до вызывающего", func(t *testing.T) {
		if _, err := r.UpdateStatus(ctx, 1, task.ID, "archived"); err == nil {
			t.Errorf("UpdateStatus(archived) err = nil: ошибку базы проглотили")
		}
	})

	t.Run("инъекция в статус фильтра безопасна", func(t *testing.T) {
		got, err := r.List(ctx, 1, Filter{Status: "' OR '1'='1"})
		if err != nil || len(got) != 0 {
			t.Errorf("List(инъекция) = %d задач, %v; want 0 задач и nil", len(got), err)
		}
	})

	t.Run("кавычки в title сохраняются как есть", func(t *testing.T) {
		title := `O'Reilly; DROP TABLE tasks; --`
		created, err := r.Create(ctx, Task{OwnerID: 1, Title: title})
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		rows := testdb.Rows(t, pool, `SELECT title FROM tasks WHERE id = $1`, created.ID)
		if len(rows) != 1 || rows[0][0] != title {
			t.Errorf("в базе title = %v, want %q", rows, title)
		}
		if !strings.Contains(created.Title, "O'Reilly") {
			t.Errorf("Create вернул title %q", created.Title)
		}
	})
}
