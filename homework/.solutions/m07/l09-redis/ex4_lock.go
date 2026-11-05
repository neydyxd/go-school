package l09redis

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/redis/go-redis/v9"
)

// TryLock пытается взять распределённую блокировку key на время ttl.
func TryLock(ctx context.Context, rdb *redis.Client, key string, ttl time.Duration) (token string, ok bool, err error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", false, err
	}
	t := hex.EncodeToString(b)
	ok, err = rdb.SetNX(ctx, key, t, ttl).Result() // SET key t PX ttl NX
	if err != nil || !ok {
		return "", false, err
	}
	return t, true, nil
}

// Lua-скрипт выполняется в Redis целиком, никто не вклинится между GET и DEL.
var unlockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0`)

// Unlock снимает блокировку, только если она всё ещё наша (значение == token).
func Unlock(ctx context.Context, rdb *redis.Client, key, token string) (bool, error) {
	n, err := unlockScript.Run(ctx, rdb, []string{key}, token).Int()
	return n == 1, err
}
