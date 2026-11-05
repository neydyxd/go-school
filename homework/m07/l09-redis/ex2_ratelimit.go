package l09redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Allow это rate limiter с фиксированным окном: не больше limit запросов
// по ключу key за окно window. Возвращает true, если запрос разрешён.
//
// Алгоритм: INCR key. Получилось <= limit: разрешить, иначе отказать.
// У ключа обязан быть TTL = window, иначе счётчик никогда не обнулится.
//
//   - TTL ставится один раз, на первом запросе окна. Если продлевать его
//     на каждом запросе, активный клиент никогда не дождётся нового окна;
//   - INCR и установка TTL должны быть надёжны вместе: если процесс упадёт
//     между ними, ключ без TTL заблокирует клиента навсегда. Отправь обе
//     команды одной транзакцией (TxPipelined) и используй EXPIRE ... NX
//     (в go-redis: rdb.ExpireNX), который ставит TTL, только если его ещё нет;
//   - конкурентные вызовы не должны пропускать больше limit запросов.
func Allow(ctx context.Context, rdb *redis.Client, key string, limit int, window time.Duration) (bool, error) {
	panic("TODO: Allow")
}
