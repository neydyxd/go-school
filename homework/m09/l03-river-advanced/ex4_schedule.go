package l03riveradvanced

import (
	"time"

	"github.com/riverqueue/river"
)

// DailyAt это расписание «каждый день в Hour:Minute по часам Loc». Оно
// реализует river.PeriodicSchedule, поэтому годится для river.NewPeriodicJob.
type DailyAt struct {
	Hour, Minute int
	Loc          *time.Location
}

var _ river.PeriodicSchedule = DailyAt{}

// Next возвращает ближайший момент СТРОГО после current, когда часы в Loc
// показывают Hour:Minute:00. Ровно в Hour:Minute следующий запуск через сутки.
//
// Сутки считай по календарю Loc (time.Date с day+1), а не прибавлением
// 24*time.Hour: в день перевода часов сутки длятся 23 или 25 часов, и запуск
// «съедет» на час. current может прийти в любом часовом поясе, переведи его в Loc.
func (d DailyAt) Next(current time.Time) time.Time {
	return time.Time{} // TODO
}

// OverdueDigestArgs: периодическая задача «разослать дайджест просроченного».
type OverdueDigestArgs struct{}

func (OverdueDigestArgs) Kind() string { return "overdue_digest" }

// NewDigestJob: дайджест каждый день в 09:00 по Москве. Готово, это образец
// того, как расписание подключается к клиенту: river.Config{PeriodicJobs: ...}.
func NewDigestJob(moscow *time.Location) *river.PeriodicJob {
	return river.NewPeriodicJob(
		DailyAt{Hour: 9, Minute: 0, Loc: moscow},
		func() (river.JobArgs, *river.InsertOpts) {
			// уникальность на сутки страхует от двойного запуска при смене лидера
			return OverdueDigestArgs{}, &river.InsertOpts{UniqueOpts: river.UniqueOpts{ByPeriod: 24 * time.Hour}}
		},
		&river.PeriodicJobOpts{ID: "overdue_digest"},
	)
}
