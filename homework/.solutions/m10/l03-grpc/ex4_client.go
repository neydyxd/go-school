package l03grpc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	resp, err := client.GetTask(ctx, &tasksv1.GetTaskRequest{Id: id})
	switch status.Code(err) {
	case codes.OK:
		return resp.GetTask(), nil
	case codes.NotFound:
		return nil, ErrNotFound
	case codes.DeadlineExceeded:
		return nil, ErrTimeout
	default:
		return nil, fmt.Errorf("fetch task %d: %w", id, err)
	}
}
