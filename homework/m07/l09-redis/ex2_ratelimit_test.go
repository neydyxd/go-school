package l09redis

import (
	"testing"
	"time"

	"goschool/homework/internal/testredis"
)

func TestAllow(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	ctx := testCtx(t)
	key := prefix + "rl:user:1"

	for i := 1; i <= 7; i++ {
		got, err := Allow(ctx, rdb, key, 5, time.Minute)
		if err != nil {
			t.Fatalf("Allow вызов %d: %v", i, err)
		}
		if want := i <= 5; got != want {
			t.Fatalf("Allow вызов %d при лимите 5 = %v, want %v", i, got, want)
		}
	}
	ttl := rdb.PTTL(ctx, key).Val()
	if ttl <= 0 || ttl > time.Minute {
		t.Fatalf("PTTL счётчика = %v, want от 0 до 1m: без TTL окно никогда не кончится", ttl)
	}
	// у другого пользователя свой счётчик
	if ok, _ := Allow(ctx, rdb, prefix+"rl:user:2", 5, time.Minute); !ok {
		t.Errorf("первый запрос user:2 отклонён: ключи пользователей должны быть независимы")
	}
}

// Фиксированное окно: TTL ставится один раз и не продлевается.
func TestAllowFixedWindow(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	ctx := testCtx(t)
	key := prefix + "rl:fixed"

	if ok, err := Allow(ctx, rdb, key, 100, time.Minute); !ok || err != nil {
		t.Fatalf("первый Allow = (%v, %v), want (true, nil)", ok, err)
	}
	// имитируем, что от окна осталось 10 секунд
	rdb.PExpire(ctx, key, 10*time.Second)
	if ok, err := Allow(ctx, rdb, key, 100, time.Minute); !ok || err != nil {
		t.Fatalf("второй Allow = (%v, %v), want (true, nil)", ok, err)
	}
	if ttl := rdb.PTTL(ctx, key).Val(); ttl > 10*time.Second {
		t.Errorf("PTTL после второго запроса = %v, want <= 10s: TTL продлевается на каждом запросе, окно не кончится никогда", ttl)
	}
}

func TestAllowConcurrent(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	const n, limit = 50, 10
	results := make(chan bool, n)
	for range n {
		go func() {
			ok, err := Allow(testCtx(t), rdb, prefix+"rl:burst", limit, time.Minute)
			if err != nil {
				t.Error(err)
			}
			results <- ok
		}()
	}
	allowed := 0
	for range n {
		if <-results {
			allowed++
		}
	}
	if allowed != limit {
		t.Errorf("из %d одновременных запросов пропущено %d, want %d", n, allowed, limit)
	}
}
