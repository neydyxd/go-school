package l06patterns

import (
	"context"

	"golang.org/x/sync/errgroup"
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
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(max(limit, 1))
	out := make([]string, len(ids))
	for i, id := range ids {
		g.Go(func() error {
			if err := ctx.Err(); err != nil {
				return err
			}
			s, err := fetch(ctx, id)
			if err != nil {
				return err
			}
			out[i] = s
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return out, nil
}
