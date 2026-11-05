package l09redis

import (
	"context"
	"encoding/json"
)

// GetOnce это Get, защищённый от cache stampede (эффект толпы).
func (c *ProductCache) GetOnce(ctx context.Context, id int64) (Product, error) {
	if p, ok := c.cached(ctx, id); ok {
		return p, nil
	}
	// Все одновременные промахи по ключу ждут один вызов. Внутри c.Get
	// снова читает кеш: опоздавший «полёт» найдёт значение и в базу не пойдёт.
	// ponytail: ctx первого вызова общий для всех ждущих; нужна независимость, бери context.WithoutCancel
	v, err, _ := c.group.Do(c.key(id), func() (any, error) {
		return c.Get(ctx, id)
	})
	if err != nil {
		return Product{}, err
	}
	return v.(Product), nil
}

// cached читает товар из кеша без похода в источник.
func (c *ProductCache) cached(ctx context.Context, id int64) (Product, bool) {
	raw, err := c.rdb.Get(ctx, c.key(id)).Bytes()
	if err != nil {
		return Product{}, false
	}
	var p Product
	return p, json.Unmarshal(raw, &p) == nil
}
