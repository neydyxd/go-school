package jobs

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// TaskAssignedWorker отправляет исполнителю письмо о назначении задачи. TODO.
//
// Work:
//  1. SELECT t.title, t.assignee_id, u.email FROM tasks t
//     JOIN users u ON u.id = $2 WHERE t.id = $1.
//     Задачи или пользователя нет: river.JobCancel(...), повторять бессмысленно.
//  2. t.assignee_id уже не равен job.Args.AssigneeID (задачу успели
//     переназначить): верни nil, письмо устарело, новое поставит новое назначение.
//  3. Ключ идемпотентности: fmt.Sprintf("task_assigned:%d", job.ID). Он один на
//     все попытки задачи. Есть строка в sent_emails с этим ключом: верни nil.
//  4. Mailer.Send(ctx, Email{To: email, Subject: "Тебе назначили задачу", Body:
//     title, IdempotencyKey: key}). Ошибку верни: River повторит.
//  5. INSERT INTO sent_emails (idempotency_key) VALUES (key) ON CONFLICT DO NOTHING.
//
// Тесты: rivertest.NewWorker + FakeMailer (см. ТЗ в уроке m09/l03).
type TaskAssignedWorker struct {
	river.WorkerDefaults[TaskAssignedArgs]

	Pool   *pgxpool.Pool
	Mailer Mailer
}

func (w *TaskAssignedWorker) Work(ctx context.Context, job *river.Job[TaskAssignedArgs]) error {
	return ErrNotImplemented
}

// OverdueDigestWorker рассылает дайджест просроченных задач. TODO.
//
// Work:
//  1. Одним запросом выбери просроченные задачи: due_at < now(), status <> 'done',
//     assignee_id IS NOT NULL, с email исполнителя, ORDER BY assignee_id, due_at.
//  2. Сгруппируй по исполнителю: одно письмо на человека со списком задач.
//  3. Ключ идемпотентности: "digest:<assignee_id>:<дата UTC 2006-01-02>".
//     Уже есть в sent_emails: этому человеку сегодня не шли. Иначе Send и INSERT
//     в sent_emails, как в TaskAssignedWorker.
//  4. Ошибка отправки одному человеку не должна отменять письма остальным:
//     отправь всем, кому получится, и верни errors.Join из неудач. При повторе
//     задачи уже отправленные пропустятся по ключу.
//
// Со звёздочкой: при тысячах исполнителей вместо писем прямо здесь поставь по
// задаче на человека через InsertMany (fan-out) и отправляй их в очереди email.
type OverdueDigestWorker struct {
	river.WorkerDefaults[OverdueDigestArgs]

	Pool   *pgxpool.Pool
	Mailer Mailer
}

func (w *OverdueDigestWorker) Work(ctx context.Context, job *river.Job[OverdueDigestArgs]) error {
	return ErrNotImplemented
}
