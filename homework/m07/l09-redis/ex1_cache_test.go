package l09redis

import (
	"errors"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"goschool/homework/internal/testredis"
)

func TestCacheAside(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	ctx := testCtx(t)
	db := newFakeDB()
	c := NewProductCache(rdb, prefix, time.Minute, db.load)

	want := Product{ID: 42, Name: "Кружка", Price: 59900}
	for i := range 3 {
		got, err := c.Get(ctx, 42)
		if err != nil || got != want {
			t.Fatalf("Get(42) вызов %d = (%+v, %v), want (%+v, nil)", i+1, got, err, want)
		}
	}
	if n := db.calls.Load(); n != 1 {
		t.Fatalf("три Get(42) сходили в базу %d раз, want 1: второй и третий должны попасть в кеш", n)
	}

	key := prefix + "product:42"
	ttl, err := rdb.PTTL(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	if ttl <= 0 || ttl > time.Minute {
		t.Errorf("PTTL %s = %v, want от 0 до 1m: ключ в кеше обязан иметь TTL", key, ttl)
	}

	// товар поменялся в базе: кеш отдаёт старое, пока не инвалидируем
	db.set(Product{ID: 42, Name: "Кружка", Price: 49900})
	if got, _ := c.Get(ctx, 42); got.Price != 59900 {
		t.Fatalf("до Invalidate Get(42).Price = %d, want 59900 из кеша", got.Price)
	}
	if err := c.Invalidate(ctx, 42); err != nil {
		t.Fatalf("Invalidate(42) = %v", err)
	}
	if n := rdb.Exists(ctx, key).Val(); n != 0 {
		t.Fatalf("после Invalidate ключ %s остался в Redis", key)
	}
	if got, err := c.Get(ctx, 42); err != nil || got.Price != 49900 {
		t.Fatalf("после Invalidate Get(42) = (%+v, %v), want цену 49900", got, err)
	}
	if n := db.calls.Load(); n != 2 {
		t.Errorf("обращений к базе %d, want 2", n)
	}
}

func TestCacheAsideErrorsNotCached(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	ctx := testCtx(t)
	db := newFakeDB()
	c := NewProductCache(rdb, prefix, time.Minute, db.load)

	for i := range 2 {
		if _, err := c.Get(ctx, 7); !errors.Is(err, ErrNotFound) {
			t.Fatalf("Get(7) вызов %d = %v, want ErrNotFound", i+1, err)
		}
	}
	if n := db.calls.Load(); n != 2 {
		t.Errorf("Get(7) дважды сходил в базу %d раз, want 2: ошибку нельзя класть в кеш", n)
	}
	if n := rdb.Exists(ctx, prefix+"product:7").Val(); n != 0 {
		t.Errorf("для несуществующего товара в Redis появился ключ")
	}
	// товар появился: следующий Get его находит
	db.set(Product{ID: 7, Name: "Ложка", Price: 9900})
	if got, err := c.Get(ctx, 7); err != nil || got.Name != "Ложка" {
		t.Errorf("Get(7) после появления товара = (%+v, %v)", got, err)
	}
}

func TestCacheAsideGarbage(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	ctx := testCtx(t)
	db := newFakeDB()
	c := NewProductCache(rdb, prefix, time.Minute, db.load)

	rdb.Set(ctx, prefix+"product:1", "{не json", time.Minute)
	got, err := c.Get(ctx, 1)
	if err != nil || got.Name != "Чайник" {
		t.Fatalf("Get(1) при мусоре в кеше = (%+v, %v), want товар из базы", got, err)
	}
	if _, err := c.Get(ctx, 1); err != nil || db.calls.Load() != 1 {
		t.Errorf("после перезаписи мусора второй Get(1) снова пошёл в базу (обращений %d), want 1", db.calls.Load())
	}
}

func TestCacheAsideRedisDown(t *testing.T) {
	t.Parallel()
	// порт 1 никто не слушает: Redis «упал»
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:1", MaxRetries: -1, DialTimeout: 200 * time.Millisecond})
	t.Cleanup(func() { rdb.Close() })
	db := newFakeDB()
	c := NewProductCache(rdb, "t:down:", time.Minute, db.load)

	got, err := c.Get(testCtx(t), 42)
	if err != nil || got.Name != "Кружка" {
		t.Fatalf("Get(42) без Redis = (%+v, %v), want товар из базы без ошибки", got, err)
	}
}
