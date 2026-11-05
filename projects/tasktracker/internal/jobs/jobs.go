// Package jobs: фоновые задачи трекера на River (модуль 09).
//
// Здесь аргументы задач, воркеры, интерфейс отправки писем и сборка клиента.
// Ставит задачи сервис через service.Store.Jobs в той же транзакции, что и
// бизнес-запись (реализация в repository.Jobs, InsertTx).
package jobs

import (
	"errors"
	"time"

	"github.com/riverqueue/river"
)

// QueueEmail это очередь для писем: у почтового провайдера свои лимиты, поэтому
// письма не делят воркеры с остальными задачами.
const QueueEmail = "email"

// ErrNotImplemented возвращают заготовки воркеров. River будет повторять такие
// задачи по backoff, пока ты не напишешь Work: ничего не потеряется.
var ErrNotImplemented = errors.New("воркер ещё не написан")

// TaskAssignedArgs: «задачу TaskID назначили на пользователя AssigneeID,
// отправь ему письмо». Kind не менять после выкатки.
type TaskAssignedArgs struct {
	TaskID     int64 `json:"task_id"`
	AssigneeID int64 `json:"assignee_id"`
}

func (TaskAssignedArgs) Kind() string { return "task_assigned" }

func (TaskAssignedArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{Queue: QueueEmail, MaxAttempts: 10}
}

// OverdueDigestArgs: «разошли исполнителям письма со списком их просроченных
// задач». Ставится периодически (DigestPeriodicJob), не чаще раза в сутки.
type OverdueDigestArgs struct{}

func (OverdueDigestArgs) Kind() string { return "overdue_digest" }

// DigestPeriodicJob: дайджест раз в сутки. RunOnStart ставит задачу при старте
// лидера, а уникальность по суткам (UTC) не даёт двух дайджестов в день, сколько
// бы раз сервис ни перезапускали. Готово.
//
// Со звёздочкой: замени PeriodicInterval на расписание «каждый день в 09:00 по
// Москве» (DailyAt из домашки m09/l03).
func DigestPeriodicJob() *river.PeriodicJob {
	return river.NewPeriodicJob(
		river.PeriodicInterval(24*time.Hour),
		func() (river.JobArgs, *river.InsertOpts) {
			return OverdueDigestArgs{}, &river.InsertOpts{
				UniqueOpts: river.UniqueOpts{ByPeriod: 24 * time.Hour},
			}
		},
		&river.PeriodicJobOpts{ID: "overdue_digest", RunOnStart: true},
	)
}
