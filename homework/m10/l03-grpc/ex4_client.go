package l03grpc

import (
	"context"
	"errors"
	"time"

	tasksv1 "goschool/homework/m10/l03-grpc/gen/tasks/v1"
)

// Доменные ошибки клиента: вызывающий код не должен знать про gRPC-коды.
var (
	ErrNotFound = errors.New("задача не найдена")
	ErrTimeout  = errors.New("сервис задач не ответил вовремя")
)

// FetchTask получает задачу через client с дедлайном timeout (дедлайн
// уезжает на сервер вместе с запросом) и переводит gRPC-ошибки в доменные:
//
//   - codes.NotFound: ErrNotFound;
//   - codes.DeadlineExceeded: ErrTimeout;
//   - остальные ошибки: обернуть с текстом "fetch task <id>: ..." так, чтобы
//     status.Code(err) по-прежнему возвращал исходный код (%w).
func FetchTask(ctx context.Context, client tasksv1.TaskServiceClient, id int64, timeout time.Duration) (*tasksv1.Task, error) {
	panic("TODO: реализуй FetchTask")
}
