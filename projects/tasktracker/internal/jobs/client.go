package jobs

import (
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

// NewClient собирает клиент River: воркеры, очереди, периодический дайджест.
// Запуск (Start) и остановка (Stop) в cmd/api/main.go. Готово.
func NewClient(pool *pgxpool.Pool, log *slog.Logger, mailer Mailer) (*river.Client[pgx.Tx], error) {
	workers := river.NewWorkers()
	river.AddWorker(workers, &TaskAssignedWorker{Pool: pool, Mailer: mailer})
	river.AddWorker(workers, &OverdueDigestWorker{Pool: pool, Mailer: mailer})

	return river.NewClient(riverpgxv5.New(pool), &river.Config{
		Logger: log,
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 5},
			QueueEmail:         {MaxWorkers: 2},
		},
		Workers:      workers,
		PeriodicJobs: []*river.PeriodicJob{DigestPeriodicJob()},
		// при остановке даём письмам 10 секунд дописаться, потом отменяем ctx
		SoftStopTimeout: 10 * time.Second,
	})
}
