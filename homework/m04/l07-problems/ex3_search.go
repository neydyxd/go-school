package l07problems

import (
	"context"
	"errors"
)

// ErrAllFailed: ни одна реплика не ответила успешно.
var ErrAllFailed = errors.New("все реплики вернули ошибку")

// Replica ищет query и возвращает ответ.
type Replica func(ctx context.Context, query string) (string, error)

// Search отправляет query во все реплики одновременно и возвращает первый
// успешный ответ (классика из доклада Роба Пайка про паттерны конкурентности).
//   - Возвращается СРАЗУ, как пришёл первый успех, не дожидаясь остальных.
//   - Остальным репликам отменяет контекст.
//   - Все ответили ошибкой (или реплик нет): вернуть ошибку, для которой
//     errors.Is(err, ErrAllFailed) и errors.Is(err, каждая ошибка реплики).
//   - Родительский ctx отменён раньше: сразу вернуть ctx.Err().
//   - Утечек нет: даже если реплика игнорирует отмену и ответит через час,
//     её горутина после этого должна завершиться, а не повиснуть навсегда.
func Search(ctx context.Context, query string, replicas ...Replica) (string, error) {
	panic("TODO: реализуй Search")
}
