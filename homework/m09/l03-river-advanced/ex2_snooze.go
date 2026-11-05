package l03riveradvanced

import (
	"context"
	"time"

	"github.com/riverqueue/river"
)

// ExportArgs: задача «забрать готовую выгрузку у внешнего сервиса».
type ExportArgs struct {
	ExportID string `json:"export_id"`
}

func (ExportArgs) Kind() string { return "export_download" }

// Статусы выгрузки у внешнего сервиса.
const (
	ExportPending = "pending" // ещё готовится
	ExportReady   = "ready"   // можно скачивать
	ExportFailed  = "failed"  // сервис сдался, ждать бесполезно
)

// ExportWorker ждёт, пока внешний сервис подготовит выгрузку, и скачивает её.
type ExportWorker struct {
	river.WorkerDefaults[ExportArgs]

	Status   func(ctx context.Context, id string) (string, error)
	Download func(ctx context.Context, id string) error
}

// Work:
//   - Status вернул ошибку: верни её (обычный повтор с backoff, попытка засчитана);
//   - ExportPending: river.JobSnooze(30 * time.Second). Это не ошибка: задача
//     уснёт на 30 секунд и не потратит попытку, сколько бы раз ни засыпала;
//   - ExportFailed или неизвестный статус: river.JobCancel(ошибка с id и статусом),
//     повторять бессмысленно;
//   - ExportReady: верни результат Download.
func (w *ExportWorker) Work(ctx context.Context, job *river.Job[ExportArgs]) error {
	panic("TODO: Work")
}

// Timeout ограничивает одну попытку двумя минутами: по истечении River отменит
// ctx, переданный в Work. По умолчанию действует JobTimeout клиента (1 минута).
func (w *ExportWorker) Timeout(job *river.Job[ExportArgs]) time.Duration {
	return 0 // TODO
}
