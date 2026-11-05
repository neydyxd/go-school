package l09redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// TryLock пытается взять распределённую блокировку key на время ttl.
//
//   - одна команда SET key token NX PX ttl (в go-redis: rdb.SetNX(ctx, key, token, ttl)).
//     Отдельные SETNX и EXPIRE не годятся: упадёшь между ними, и лок вечный;
//   - token это случайная строка (crypto/rand, 16 байт в hex): по ней Unlock
//     отличит свой лок от чужого;
//   - лок занят: ok == false, token пустой, err == nil.
func TryLock(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration) (token string, ok bool, err error) {
	panic("TODO: TryLock")
}

// Unlock снимает блокировку, только если она всё ещё наша (значение == token).
// Возвращает true, если ключ удалён.
//
// Сценарий, от которого защищаемся: наш лок истёк по TTL, его взял другой
// процесс, а мы по старой памяти делаем DEL и снимаем ЧУЖОЙ лок.
// Проверка и удаление должны быть атомарны: GET и DEL отдельными командами
// оставляют окно для гонки. Используй Lua-скрипт (redis.NewScript):
//
//	if redis.call("GET", KEYS[1]) == ARGV[1] then
//	    return redis.call("DEL", KEYS[1])
//	end
//	return 0
func Unlock(ctx context.Context, rdb *redis.Client, key, token string) (bool, error) {
	panic("TODO: Unlock")
}
