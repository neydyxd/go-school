package l03riveradvanced

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"
	"github.com/riverqueue/river/rivertype"
)

func TestExportWorker(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	errAPI := errors.New("502 от сервиса выгрузок")

	tests := []struct {
		name      string
		status    string
		statusErr error
		wantEvent river.EventKind
		wantState rivertype.JobState
		wantErr   bool
		wantDL    int
	}{
		{"готова: скачиваем", ExportReady, nil, river.EventKindJobCompleted, rivertype.JobStateCompleted, false, 1},
		{"готовится: засыпаем", ExportPending, nil, river.EventKindJobSnoozed, rivertype.JobStateScheduled, false, 0},
		{"сервис сдался: отмена", ExportFailed, nil, river.EventKindJobCancelled, rivertype.JobStateCancelled, false, 0},
		{"неизвестный статус: отмена", "archived", nil, river.EventKindJobCancelled, rivertype.JobStateCancelled, false, 0},
		{"ошибка API: обычный повтор", "", errAPI, river.EventKindJobFailed, rivertype.JobStateAvailable, true, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var downloads int
			w := &ExportWorker{
				Status: func(context.Context, string) (string, error) { return tt.status, tt.statusErr },
				Download: func(_ context.Context, id string) error {
					if id != "exp-1" {
						t.Errorf("Download(%q), want exp-1", id)
					}
					downloads++
					return nil
				},
			}
			tw := rivertest.NewWorker(t, riverpgxv5.New(nil), &river.Config{}, w)
			before := time.Now()
			res, err := tw.Work(ctx, t, beginTx(t, pool), ExportArgs{ExportID: "exp-1"}, nil)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Work error = %v, want ошибку: %v", err, tt.wantErr)
			}
			if res.EventKind != tt.wantEvent || res.Job.State != tt.wantState {
				t.Errorf("событие %s, состояние %s, want %s, %s", res.EventKind, res.Job.State, tt.wantEvent, tt.wantState)
			}
			if downloads != tt.wantDL {
				t.Errorf("Download вызван %d раз, want %d", downloads, tt.wantDL)
			}
			if tt.status == ExportPending {
				if !near(res.Job.ScheduledAt, before, 30*time.Second) {
					t.Errorf("snooze на %v, want ~30s", res.Job.ScheduledAt.Sub(before).Round(time.Second))
				}
				if res.Job.Attempt != 0 {
					t.Errorf("после snooze Attempt = %d, want 0: snooze не тратит попытку", res.Job.Attempt)
				}
			}
		})
	}
}

// MaxAttempts 1: обычная ошибка сразу отправила бы задачу в discarded,
// а snooze можно повторять сколько угодно.
func TestExportWorkerSnoozeDoesNotDiscard(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	w := &ExportWorker{
		Status:   func(context.Context, string) (string, error) { return ExportPending, nil },
		Download: func(context.Context, string) error { return nil },
	}
	tw := rivertest.NewWorker(t, riverpgxv5.New(nil), &river.Config{}, w)
	tx := beginTx(t, pool)
	res, err := tw.Work(ctx, t, tx, ExportArgs{ExportID: "exp-2"}, &river.InsertOpts{MaxAttempts: 1})
	if err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if _, err := tx.Exec(ctx, `UPDATE river_job SET state = 'available', scheduled_at = now() WHERE id = $1`, res.Job.ID); err != nil {
			t.Fatal(err)
		}
		if res, err = tw.WorkJob(ctx, t, tx, res.Job); err != nil {
			t.Fatal(err)
		}
	}
	if res.Job.State != rivertype.JobStateScheduled {
		t.Errorf("после 4 snooze при MaxAttempts 1 состояние %s, want scheduled", res.Job.State)
	}
}

func TestExportWorkerTimeout(t *testing.T) {
	if got := (&ExportWorker{}).Timeout(&river.Job[ExportArgs]{}); got != 2*time.Minute {
		t.Errorf("Timeout() = %v, want 2m", got)
	}
}
