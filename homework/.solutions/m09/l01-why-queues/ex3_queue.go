package l01whyqueues

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// DBTX это общее у *pgxpool.Pool, *pgx.Conn и pgx.Tx.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// Job это задача из таблицы queue_jobs.
type Job struct {
	ID       int64
	Kind     string
	Payload  string
	Attempts int // сколько раз задачу уже забирали, включая текущий
}

// ErrNotRunning: задачи нет или она не в состоянии 'running'.
var ErrNotRunning = errors.New("задача не в работе")

// Enqueue кладёт задачу в очередь, она станет доступна в момент runAt. Готово.
func Enqueue(ctx context.Context, db DBTX, kind, payload string, runAt time.Time) (int64, error) {
	var id int64
	err := db.QueryRow(ctx,
		`INSERT INTO queue_jobs (kind, payload, run_at) VALUES ($1, $2, $3) RETURNING id`,
		kind, payload, runAt).Scan(&id)
	return id, err
}

// Claim забирает одну задачу для воркера worker: самую раннюю по run_at среди
// state = 'available' и run_at <= now(), при равных run_at с меньшим id.
// Забранная задача получает state = 'running', locked_by = worker, attempts + 1.
// Вернуть (задача, true, nil); задач нет: (Job{}, false, nil).
//
// Главное требование: два воркера, которые зовут Claim одновременно (каждый в
// своей транзакции), никогда не получают одну задачу и НЕ ждут друг друга.
// Если строку уже заблокировала чужая незакоммиченная транзакция, бери следующую.
// Один запрос: UPDATE ... WHERE id = (SELECT id ... FOR UPDATE SKIP LOCKED LIMIT 1) RETURNING ...
func Claim(ctx context.Context, db DBTX, worker string) (Job, bool, error) {
	var j Job
	err := db.QueryRow(ctx, `
		UPDATE queue_jobs SET state = 'running', locked_by = $1, attempts = attempts + 1
		WHERE id = (
			SELECT id FROM queue_jobs
			WHERE state = 'available' AND run_at <= now()
			ORDER BY run_at, id
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		)
		RETURNING id, kind, payload, attempts`, worker).Scan(&j.ID, &j.Kind, &j.Payload, &j.Attempts)
	if errors.Is(err, pgx.ErrNoRows) {
		return Job{}, false, nil
	}
	if err != nil {
		return Job{}, false, err
	}
	return j, true, nil
}

// Complete помечает задачу выполненной: state = 'done', locked_by = NULL.
// Только если она сейчас 'running', иначе ErrNotRunning.
func Complete(ctx context.Context, db DBTX, id int64) error {
	tag, err := db.Exec(ctx,
		`UPDATE queue_jobs SET state = 'done', locked_by = NULL WHERE id = $1 AND state = 'running'`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotRunning
	}
	return nil
}

// Fail записывает неудачную попытку: last_error = errMsg, locked_by = NULL и
//   - attempts < maxAttempts: state = 'available', run_at = retryAt (повтор позже);
//   - attempts >= maxAttempts: state = 'dead' (попытки кончились, разбирать руками).
//
// Только для задачи в 'running', иначе ErrNotRunning.
func Fail(ctx context.Context, db DBTX, id int64, errMsg string, retryAt time.Time, maxAttempts int) error {
	tag, err := db.Exec(ctx, `
		UPDATE queue_jobs SET
			last_error = $2,
			locked_by = NULL,
			state = CASE WHEN attempts >= $4 THEN 'dead' ELSE 'available' END,
			run_at = CASE WHEN attempts >= $4 THEN run_at ELSE $3 END
		WHERE id = $1 AND state = 'running'`, id, errMsg, retryAt, maxAttempts)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotRunning
	}
	return nil
}
