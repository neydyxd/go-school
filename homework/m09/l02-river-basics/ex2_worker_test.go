package l02riverbasics

import (
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/riverqueue/river/rivertype"

	"goschool/homework/internal/testdb"
)

// rivertest.Worker вставляет задачу в переданную транзакцию и выполняет её
// синхронно, прямо в тесте: ни клиента, ни ожиданий. Бизнес-данные (users)
// воркер читает через пул, поэтому пользователей создаём через пул.
func newTestWorker(t *testing.T, w *WelcomeEmailWorker) *rivertest.Worker[WelcomeEmailArgs, pgx.Tx] {
	t.Helper()
	return rivertest.NewWorker(t, riverpgxv5.New(nil), &river.Config{}, w)
}

func TestWelcomeEmailWorkerIdempotent(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	mailer := &fakeMailer{}
	tw := newTestWorker(t, &WelcomeEmailWorker{Pool: pool, Mailer: mailer})
	id := addUser(t, pool, "anna@example.com")
	args := WelcomeEmailArgs{UserID: id, Email: "anna@example.com"}

	for run := 1; run <= 2; run++ { // вторая доставка той же задачи: at-least-once
		res, err := tw.Work(ctx, t, beginTx(t, pool), args, nil)
		if err != nil || res.EventKind != river.EventKindJobCompleted {
			t.Fatalf("запуск %d: Work = (%v, %v), want (job_completed, nil)", run, res.EventKind, err)
		}
	}
	if got, want := mailer.Sent(), []string{"anna@example.com|" + WelcomeSubject}; !reflect.DeepEqual(got, want) {
		t.Errorf("отправлено %v, want %v (одно письмо на два запуска)", got, want)
	}
	if got := testdb.Rows(t, pool, `SELECT welcome_sent_at IS NOT NULL FROM users WHERE id = $1`, id); got[0][0] != "true" {
		t.Error("welcome_sent_at не заполнен после отправки")
	}
}

func TestWelcomeEmailWorkerUserGone(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	mailer := &fakeMailer{}
	tw := newTestWorker(t, &WelcomeEmailWorker{Pool: pool, Mailer: mailer})

	res, err := tw.Work(ctx, t, beginTx(t, pool), WelcomeEmailArgs{UserID: 404, Email: "ghost@example.com"}, nil)
	if err != nil {
		t.Fatalf("Work вернула ошибку %v: JobCancel не считается ошибкой rivertest, а обычная ошибка считается", err)
	}
	if res.EventKind != river.EventKindJobCancelled || res.Job.State != rivertype.JobStateCancelled {
		t.Errorf("удалённый пользователь: событие %s, состояние %s, want job_cancelled, cancelled", res.EventKind, res.Job.State)
	}
	if len(mailer.Sent()) != 0 {
		t.Errorf("отправлено %v, want ничего", mailer.Sent())
	}
}

func TestWelcomeEmailWorkerRetryAfterFailure(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	mailer := &fakeMailer{failNext: 1}
	tw := newTestWorker(t, &WelcomeEmailWorker{Pool: pool, Mailer: mailer})
	id := addUser(t, pool, "boris@example.com")
	args := WelcomeEmailArgs{UserID: id, Email: "boris@example.com"}

	res, err := tw.Work(ctx, t, beginTx(t, pool), args, nil)
	if err == nil || res.EventKind != river.EventKindJobFailed {
		t.Fatalf("SMTP ответил ошибкой: Work = (%v, %v), want (job_failed, ошибку)", res.EventKind, err)
	}
	if res.Job.State == rivertype.JobStateCancelled || res.Job.State == rivertype.JobStateDiscarded {
		t.Errorf("после временной ошибки задача в состоянии %s, want ожидание повтора", res.Job.State)
	}
	if got := testdb.Rows(t, pool, `SELECT welcome_sent_at IS NULL FROM users WHERE id = $1`, id); got[0][0] != "true" {
		t.Fatal("письмо не ушло, а welcome_sent_at уже заполнен: повтор его не отправит")
	}

	res, err = tw.Work(ctx, t, beginTx(t, pool), args, nil)
	if err != nil || res.EventKind != river.EventKindJobCompleted {
		t.Fatalf("повтор: Work = (%v, %v), want (job_completed, nil)", res.EventKind, err)
	}
	if len(mailer.Sent()) != 1 {
		t.Errorf("отправлено %v, want одно письмо", mailer.Sent())
	}
}
