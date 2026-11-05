package repository

import (
	"context"
	"errors"
	"testing"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"

	"goschool/tasktracker/internal/jobs"
	"goschool/tasktracker/internal/service"
	"goschool/tasktracker/migrations"
)

// Образец (модуль 09): задача River ставится в той же транзакции, что и
// бизнес-запись. Коммит: задача в очереди. Откат: ни записи, ни задачи.
func TestJobsInTx(t *testing.T) {
	ctx := context.Background()
	pool := newTestPool(t)
	if err := migrations.RiverUp(ctx, pool); err != nil {
		t.Fatal(err)
	}
	// клиент только для вставки: без Queues и без Start
	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		t.Fatal(err)
	}
	tm := NewTxManager(pool).WithRiver(client)
	driver := riverpgxv5.New(pool)

	// успешная транзакция: пользователь и задача в очереди вместе
	err = tm.InTx(ctx, func(s service.Store) error {
		u, err := s.Users.Create(ctx, service.User{Email: "a@example.com", PasswordHash: "h", Role: "user"})
		if err != nil {
			return err
		}
		return s.Jobs.TaskAssigned(ctx, 42, u.ID)
	})
	if err != nil {
		t.Fatalf("InTx: %v", err)
	}
	job := rivertest.RequireInserted(ctx, t, driver, &jobs.TaskAssignedArgs{},
		&rivertest.RequireInsertedOpts{Queue: jobs.QueueEmail})
	if job.Args.TaskID != 42 {
		t.Errorf("TaskID = %d, want 42", job.Args.TaskID)
	}

	// вторая операция падает (email занят): откатываются и запись, и задача
	err = tm.InTx(ctx, func(s service.Store) error {
		if err := s.Jobs.TaskAssigned(ctx, 43, 1); err != nil {
			return err
		}
		_, err := s.Users.Create(ctx, service.User{Email: "a@example.com", PasswordHash: "h", Role: "user"})
		return err
	})
	if !errors.Is(err, service.ErrConflict) {
		t.Fatalf("InTx с дублем email = %v, want ErrConflict", err)
	}
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM river_job`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Errorf("в river_job %d задач, want 1: задача из откатившейся транзакции осталась", n)
	}

	// без WithRiver Jobs честно говорит, что River не подключён
	err = NewTxManager(pool).InTx(ctx, func(s service.Store) error { return s.Jobs.TaskAssigned(ctx, 1, 1) })
	if err == nil {
		t.Error("TaskAssigned без WithRiver = nil, want ошибку")
	}
}
