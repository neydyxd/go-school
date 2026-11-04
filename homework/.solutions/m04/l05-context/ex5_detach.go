package l05context

import (
	"context"
	"time"
)

// Detach готовит контекст для фоновой работы, которая должна пережить
// запрос (например, запись аудита после ответа клиенту):
//
//   - значения из ctx (request id, пользователь) сохраняются;
//   - отмена и дедлайн ctx на результат не действуют;
//   - у результата свой таймаут d, чтобы фоновая работа не висела вечно.
//
// Вернуть контекст и функцию отмены, которую вызывающий обязан вызвать.
func Detach(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.WithoutCancel(ctx), d)
}
