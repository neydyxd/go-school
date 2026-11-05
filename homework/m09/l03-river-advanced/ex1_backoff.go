package l03riveradvanced

import (
	"context"
	"time"

	"github.com/riverqueue/river"
)

// Backoff возвращает задержку перед повтором после errorCount-й ошибки подряд
// (1 значит «упала первая попытка»). Экспонента с потолком и «равным» джиттером:
//
//	d = min(base * 2^(errorCount-1), limit)
//	результат = d/2 + rnd * d/2, где rnd из [0, 1)
//
// errorCount < 1 считай за 1. Никакого переполнения: при errorCount 64, 100,
// 1000 результат считается от limit, а не от отрицательного или нулевого числа.
// Подсказка: удваивай в цикле и останавливайся, как только d >= limit (или
// сравнивай с limit/2 перед удвоением), а не считай 1 << (n-1) напрямую.
//
// Функция чистая: случайность приходит параметром rnd, поэтому её легко тестировать.
func Backoff(errorCount int, base, limit time.Duration, rnd float64) time.Duration {
	return 0 // TODO
}

// ReportArgs: задача «собрать отчёт».
type ReportArgs struct {
	ReportID int64 `json:"report_id"`
}

func (ReportArgs) Kind() string { return "report" }

// InsertOpts: не больше 3 попыток. Готово.
func (ReportArgs) InsertOpts() river.InsertOpts { return river.InsertOpts{MaxAttempts: 3} }

// ReportWorker собирает отчёт через Build. Готово всё, кроме NextRetry.
type ReportWorker struct {
	river.WorkerDefaults[ReportArgs]

	Build func(ctx context.Context, id int64) error
	Rand  func() float64 // источник случайности для джиттера; в тестах фиксированный
}

func (w *ReportWorker) Work(ctx context.Context, job *river.Job[ReportArgs]) error {
	return w.Build(ctx, job.Args.ReportID)
}

// NextRetry переопределяет политику повторов River для этого воркера: момент
// следующей попытки = time.Now() + Backoff(n, 30*time.Second, time.Hour, w.Rand()),
// где n это номер текущей ошибки. River вызывает NextRetry ДО того, как допишет
// текущую ошибку в job.Errors, поэтому n = len(job.Errors) + 1.
func (w *ReportWorker) NextRetry(job *river.Job[ReportArgs]) time.Time {
	return time.Time{} // TODO: сейчас работает политика River по умолчанию
}
