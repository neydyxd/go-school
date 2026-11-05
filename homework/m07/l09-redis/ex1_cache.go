package l09redis

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/singleflight"
)

var ErrNotFound = errors.New("товар не найден")

type Product struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
}

// Loader достаёт товар из основного хранилища (в жизни из PostgreSQL).
type Loader func(ctx context.Context, id int64) (Product, error)

// ProductCache это кеш товаров в Redis по схеме cache-aside.
type ProductCache struct {
	rdb    *redis.Client
	prefix string
	ttl    time.Duration
	load   Loader
	group  singleflight.Group // понадобится в задаче 3
}

func NewProductCache(rdb *redis.Client, prefix string, ttl time.Duration, load Loader) *ProductCache {
	return &ProductCache{rdb: rdb, prefix: prefix, ttl: ttl, load: load}
}

// key возвращает ключ товара: <prefix>product:<id>.
func (c *ProductCache) key(id int64) string {
	return c.prefix + "product:" + strconv.FormatInt(id, 10)
}

// Get отдаёт товар по схеме cache-aside:
//  1. GET ключа c.key(id). Есть значение: это JSON товара, верни его.
//  2. Нет (ошибка redis.Nil): вызови c.load, положи результат в Redis
//     как JSON с TTL c.ttl (SET key value EX ...) и верни.
//
// Граничные случаи:
//   - ошибку загрузчика (например, ErrNotFound) верни как есть и НЕ кешируй;
//   - в кеше мусор, который не разбирается как JSON: считай это промахом,
//     загрузи товар и перезапиши ключ;
//   - Redis недоступен (любая ошибка GET, кроме redis.Nil): кеш это
//     оптимизация, а не источник правды. Загрузи товар из c.load и верни
//     без ошибки, ошибку записи в кеш тоже проглоти.
func (c *ProductCache) Get(ctx context.Context, id int64) (Product, error) {
	panic("TODO: Get")
}

// Invalidate удаляет товар из кеша. Вызывается после изменения товара
// в основном хранилище: следующий Get загрузит свежую версию.
func (c *ProductCache) Invalidate(ctx context.Context, id int64) error {
	panic("TODO: Invalidate")
}
