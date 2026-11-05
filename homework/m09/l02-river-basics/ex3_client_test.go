package l02riverbasics

import (
	"context"
	"testing"
	"time"

	"github.com/riverqueue/river"

	"goschool/homework/internal/testdb"
)

// Сквозной тест: настоящий клиент, настоящие очереди. Вместо sleep ждём событие
// клиента (Subscribe) с таймаутом.
func TestNewClientEndToEnd(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	mailer := &fakeMailer{failNext: 1} // первая попытка упадёт, River повторит через ~1 с

	client, err := NewClient(pool, mailer)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	events, cancel := client.Subscribe(river.EventKindJobCompleted, river.EventKindJobFailed, river.EventKindJobCancelled)
	defer cancel()
	if err := client.Start(ctx); err != nil {
		t.Fatalf("Start: %v", err)
	}
	defer func() {
		stopCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := client.Stop(stopCtx); err != nil {
			t.Errorf("Stop: %v", err)
		}
	}()

	tx := beginTx(t, pool)
	id, err := RegisterUserTx(ctx, tx, client, "gleb@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	var kinds []river.EventKind
	timeout := time.After(15 * time.Second)
wait:
	for {
		select {
		case ev := <-events:
			kinds = append(kinds, ev.Kind)
			if ev.Job.Queue != "email" {
				t.Errorf("задачу выполнили в очереди %q, want email", ev.Job.Queue)
			}
			if ev.Kind == river.EventKindJobCompleted {
				break wait
			}
		case <-timeout:
			t.Fatalf("задача не выполнилась за 15 секунд, события: %v. Очередь \"email\" есть в Queues? Воркер зарегистрирован?", kinds)
		}
	}
	if len(kinds) != 2 || kinds[0] != river.EventKindJobFailed {
		t.Errorf("события %v, want [job_failed job_completed]: первая попытка падает, вторая проходит", kinds)
	}
	if got := mailer.Sent(); len(got) != 1 {
		t.Errorf("отправлено %v, want одно письмо", got)
	}
	if got := testdb.Rows(t, pool, `SELECT welcome_sent_at IS NOT NULL FROM users WHERE id = $1`, id); got[0][0] != "true" {
		t.Error("welcome_sent_at не заполнен")
	}
}
