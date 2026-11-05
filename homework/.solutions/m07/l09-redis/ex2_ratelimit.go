package l09redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Allow это rate limiter с фиксированным окном: не больше limit запросов
// по ключу key за окно window. Возвращает true, если запрос разрешён.
func Allow(ctx context.Context, rdb *redis.Client, key string, limit int, window time.Duration) (bool, error) {
	var incr *redis.IntCmd
	// MULTI/EXEC: обе команды выполнятся вместе, между ними никто не вклинится.
	_, err := rdb.TxPipelined(ctx, func(pipe redis.Pipeliner) error {
		incr = pipe.Incr(ctx, key)
		pipe.ExpireNX(ctx, key, window) // TTL только у нового окна, существующий не продлеваем
		return nil
	})
	if err != nil {
		return false, err
	}
	return incr.Val() <= int64(limit), nil
}
