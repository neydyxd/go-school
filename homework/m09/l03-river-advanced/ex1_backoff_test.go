package l03riveradvanced

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/riverqueue/river/rivertype"
)

func TestBackoff(t *testing.T) {
	const s = time.Second
	tests := []struct {
		name       string
		errorCount int
		base, max  time.Duration
		rnd        float64
		want       time.Duration
	}{
		{"первая ошибка, rnd 0: половина base", 1, 10 * s, time.Hour, 0, 5 * s},
		{"первая ошибка, rnd 0.5", 1, 10 * s, time.Hour, 0.5, 7500 * time.Millisecond},
		{"вторая ошибка: удвоение", 2, 10 * s, time.Hour, 0, 10 * s},
		{"пятая ошибка: 10s * 16", 5, 10 * s, time.Hour, 0, 80 * s},
		{"упёрлись в потолок", 10, 10 * s, time.Minute, 0, 30 * s},
		{"потолок и rnd 0.75", 10, 10 * s, time.Minute, 0.75, 52500 * time.Millisecond},
		{"ноль это как первая", 0, 10 * s, time.Hour, 0, 5 * s},
		{"отрицательное это как первая", -5, 10 * s, time.Hour, 0, 5 * s},
		{"64 ошибки: без переполнения", 64, 10 * s, time.Hour, 0, 30 * time.Minute},
		{"1000 ошибок: без переполнения", 1000, 10 * s, time.Hour, 0, 30 * time.Minute},
		{"base больше потолка", 1, time.Hour, time.Minute, 0, 30 * s},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Backoff(tt.errorCount, tt.base, tt.max, tt.rnd); got != tt.want {
				t.Errorf("Backoff(%d, %v, %v, %v) = %v, want %v", tt.errorCount, tt.base, tt.max, tt.rnd, got, tt.want)
			}
		})
	}
}

// Отчёт падает три раза подряд. NextRetry обязан давать 15 с, 30 с (rnd = 0),
// а после третьей ошибки (MaxAttempts 3) задача уходит в discarded.
func TestReportWorkerNextRetry(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	errBuild := errors.New("сервис отчётов недоступен")
	w := &ReportWorker{
		Build: func(context.Context, int64) error { return errBuild },
		Rand:  func() float64 { return 0 },
	}
	tw := rivertest.NewWorker(t, riverpgxv5.New(nil), &river.Config{}, w)
	tx := beginTx(t, pool)

	before := time.Now()
	res, err := tw.Work(ctx, t, tx, ReportArgs{ReportID: 7}, nil)
	if !errors.Is(err, errBuild) || res.EventKind != river.EventKindJobFailed {
		t.Fatalf("Work = (%v, %v), want (job_failed, ошибку Build)", res.EventKind, err)
	}
	if !near(res.Job.ScheduledAt, before, 15*time.Second) {
		t.Fatalf("после 1-й ошибки повтор через %v, want ~15s (Backoff(1, 30s, 1h, 0))", res.Job.ScheduledAt.Sub(before).Round(time.Second))
	}
	if res.Job.State != rivertype.JobStateRetryable {
		t.Errorf("состояние после 1-й ошибки %s, want retryable", res.Job.State)
	}

	// вторая попытка: делаем задачу доступной прямо сейчас и выполняем её же
	res = rerun(t, ctx, tw, tx, res.Job)
	if !near(res.Job.ScheduledAt, time.Now(), 30*time.Second) {
		t.Fatalf("после 2-й ошибки повтор через %v, want ~30s: номер ошибки это len(job.Errors)+1",
			time.Until(res.Job.ScheduledAt).Round(time.Second))
	}

	res = rerun(t, ctx, tw, tx, res.Job)
	if res.Job.State != rivertype.JobStateDiscarded {
		t.Errorf("после 3-й ошибки при MaxAttempts 3 состояние %s, want discarded", res.Job.State)
	}
	if len(res.Job.Errors) != 3 {
		t.Errorf("в job.Errors %d записей, want 3", len(res.Job.Errors))
	}
}

func rerun(t *testing.T, ctx context.Context, tw *rivertest.Worker[ReportArgs, pgx.Tx], tx pgx.Tx, job *rivertype.JobRow) *rivertest.WorkResult {
	t.Helper()
	if _, err := tx.Exec(ctx, `UPDATE river_job SET state = 'available', scheduled_at = now() WHERE id = $1`, job.ID); err != nil {
		t.Fatal(err)
	}
	res, err := tw.WorkJob(ctx, t, tx, job)
	if err == nil {
		t.Fatalf("повтор задачи %d: WorkJob вернул nil, want ошибку Build", job.ID)
	}
	return res
}
