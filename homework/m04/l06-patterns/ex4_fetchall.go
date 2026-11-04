package l06patterns

import (
	"context"

	// Пустой импорт держит golang.org/x/sync в go.mod (иначе go mod tidy его
	// удалит). Когда начнёшь решать, замени на обычный импорт errgroup.
	_ "golang.org/x/sync/errgroup"
)

// FetchAll вызывает fetch для каждого id и возвращает результаты в порядке ids.
// Используй errgroup.WithContext и SetLimit (golang.org/x/sync/errgroup уже
// в go.mod).
//   - Одновременно не больше limit вызовов (limit < 1 считать 1).
//   - Первая ошибка отменяет контекст остальных вызовов; FetchAll возвращает
//     nil и именно эту первую ошибку.
//   - После ошибки новые вызовы fetch не начинаются. Осторожно: errgroup сам
//     этого не делает, g.Go запускает функцию и после отмены контекста.
//   - FetchAll возвращается только после завершения всех запущенных вызовов.
func FetchAll(ctx context.Context, ids []int, limit int, fetch func(context.Context, int) (string, error)) ([]string, error) {
	panic("TODO: реализуй FetchAll")
}
