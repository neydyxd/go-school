package l02riverbasics

import (
	"errors"
	"testing"

	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"

	"goschool/homework/internal/testdb"
)

func TestWelcomeEmailArgs(t *testing.T) {
	if got := (WelcomeEmailArgs{}).Kind(); got != "welcome_email" {
		t.Errorf(`Kind() = %q, want "welcome_email"`, got)
	}
	opts := (WelcomeEmailArgs{}).InsertOpts()
	if opts.Queue != "email" || opts.MaxAttempts != 5 {
		t.Errorf("InsertOpts() = {Queue:%q MaxAttempts:%d}, want {Queue:email MaxAttempts:5}", opts.Queue, opts.MaxAttempts)
	}
}

func TestRegisterUserTxCommit(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	client := insertOnlyClient(t, pool)

	tx := beginTx(t, pool)
	id, err := RegisterUserTx(ctx, tx, client, "  Anna@Example.COM ")
	if err != nil || id == 0 {
		t.Fatalf("RegisterUserTx = (%d, %v), want (id, nil)", id, err)
	}
	// внутри транзакции задача уже видна
	job := rivertest.RequireInsertedTx[*riverpgxv5.Driver](ctx, t, tx, &WelcomeEmailArgs{},
		&rivertest.RequireInsertedOpts{Queue: "email", MaxAttempts: 5})
	if want := (WelcomeEmailArgs{UserID: id, Email: "anna@example.com"}); *job.Args != want {
		t.Errorf("аргументы задачи = %+v, want %+v", job.Args, want)
	}
	// снаружи (через пул) до коммита задачи нет: она в незакоммиченной транзакции
	rivertest.RequireNotInserted(ctx, t, riverpgxv5.New(pool), &WelcomeEmailArgs{}, nil)

	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	rivertest.RequireInserted(ctx, t, riverpgxv5.New(pool), &WelcomeEmailArgs{}, nil)
}

func TestRegisterUserTxRollback(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	client := insertOnlyClient(t, pool)

	tx := beginTx(t, pool)
	if _, err := RegisterUserTx(ctx, tx, client, "boris@example.com"); err != nil {
		t.Fatalf("RegisterUserTx = %v", err)
	}
	// что-то дальше в обработчике пошло не так, транзакция откатилась
	if err := tx.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	rivertest.RequireNotInserted(ctx, t, riverpgxv5.New(pool), &WelcomeEmailArgs{}, nil)
	if n := testdb.Rows(t, pool, `SELECT count(*) FROM river_job`)[0][0]; n != "0" {
		t.Fatalf("после отката в river_job %s задач, want 0: используй InsertTx, а не Insert", n)
	}
	if n := testdb.Rows(t, pool, `SELECT count(*) FROM users`)[0][0]; n != "0" {
		t.Fatalf("после отката в users %s строк, want 0", n)
	}
}

func TestRegisterUserTxErrors(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	client := insertOnlyClient(t, pool)

	tx := beginTx(t, pool)
	if _, err := RegisterUserTx(ctx, tx, client, "vera@example.com"); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{"занятый email", "vera@example.com", ErrEmailTaken},
		{"занятый email в другом регистре", " VERA@example.com", ErrEmailTaken},
		{"пустой", "   ", ErrBadEmail},
		{"без собаки", "vera.example.com", ErrBadEmail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := beginTx(t, pool)
			_, err := RegisterUserTx(ctx, tx, client, tt.email)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RegisterUserTx(%q) error = %v, want %v", tt.email, err, tt.wantErr)
			}
			_ = tx.Rollback(ctx)
		})
	}
	if n := testdb.Rows(t, pool, `SELECT count(*) FROM river_job`)[0][0]; n != "1" {
		t.Errorf("в river_job %s задач, want 1 (только от успешной регистрации)", n)
	}
}
