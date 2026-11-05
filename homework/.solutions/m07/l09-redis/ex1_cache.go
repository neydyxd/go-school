package l09redis

import (
	"context"
	"encoding/json"
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

// Get отдаёт товар по схеме cache-aside.
func (c *ProductCache) Get(ctx context.Context, id int64) (Product, error) {
	key := c.key(id)
	raw, err := c.rdb.Get(ctx, key).Bytes()
	if err == nil {
		var p Product
		if json.Unmarshal(raw, &p) == nil {
			return p, nil
		}
		// мусор в кеше: идём как при промахе и перезаписываем
	} else if !errors.Is(err, redis.Nil) {
		// Redis лежит: отдаём из источника и даже не пытаемся писать в кеш
		return c.load(ctx, id)
	}

	p, err := c.load(ctx, id)
	if err != nil {
		return Product{}, err // ошибки не кешируем
	}
	if b, err := json.Marshal(p); err == nil {
		_ = c.rdb.Set(ctx, key, b, c.ttl).Err() // ponytail: не записали в кеш, значит в следующий раз снова промах
	}
	return p, nil
}

// Invalidate удаляет товар из кеша.
func (c *ProductCache) Invalidate(ctx context.Context, id int64) error {
	return c.rdb.Del(ctx, c.key(id)).Err()
}
