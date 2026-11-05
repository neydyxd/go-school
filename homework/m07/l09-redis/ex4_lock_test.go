package l09redis

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"goschool/homework/internal/testredis"
)

func TestLock(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	ctx := testCtx(t)
	key := prefix + "lock:report"

	tokA, ok, err := TryLock(ctx, rdb, key, 30*time.Second)
	if err != nil || !ok || tokA == "" {
		t.Fatalf("первый TryLock = (%q, %v, %v), want (токен, true, nil)", tokA, ok, err)
	}
	if ttl := rdb.PTTL(ctx, key).Val(); ttl <= 0 || ttl > 30*time.Second {
		t.Fatalf("PTTL лока = %v, want от 0 до 30s: лок без TTL вечный, если владелец упадёт", ttl)
	}
	if tok, ok, err := TryLock(ctx, rdb, key, 30*time.Second); ok || err != nil || tok != "" {
		t.Fatalf("второй TryLock при занятом локе = (%q, %v, %v), want (\"\", false, nil)", tok, ok, err)
	}
	if ok, err := Unlock(ctx, rdb, key, "чужой-токен"); ok || err != nil {
		t.Fatalf("Unlock с чужим токеном = (%v, %v), want (false, nil)", ok, err)
	}
	if rdb.Exists(ctx, key).Val() != 1 {
		t.Fatalf("Unlock с чужим токеном удалил лок")
	}

	// Лок A истёк (имитируем удалением), его взял B.
	rdb.Del(ctx, key)
	tokB, ok, err := TryLock(ctx, rdb, key, 30*time.Second)
	if err != nil || !ok {
		t.Fatalf("TryLock после истечения = (%v, %v), want (true, nil)", ok, err)
	}
	if tokB == tokA {
		t.Fatalf("два лока получили одинаковый токен %q: токен должен быть случайным", tokA)
	}
	// A проснулся и пытается снять «свой» лок
	if ok, err := Unlock(ctx, rdb, key, tokA); ok || err != nil {
		t.Fatalf("Unlock старым токеном A = (%v, %v), want (false, nil): снят чужой лок", ok, err)
	}
	if rdb.Get(ctx, key).Val() != tokB {
		t.Fatalf("лок B пропал после Unlock со старым токеном A")
	}
	if ok, err := Unlock(ctx, rdb, key, tokB); !ok || err != nil {
		t.Fatalf("Unlock своим токеном B = (%v, %v), want (true, nil)", ok, err)
	}
	if rdb.Exists(ctx, key).Val() != 0 {
		t.Fatalf("после Unlock ключ остался")
	}
}

func TestLockConcurrent(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	const n = 30
	wins := make(chan bool, n)
	for range n {
		go func() {
			_, ok, err := TryLock(testCtx(t), rdb, prefix+"lock:cron", time.Minute)
			if err != nil {
				t.Error(err)
			}
			wins <- ok
		}()
	}
	won := 0
	for range n {
		if <-wins {
			won++
		}
	}
	if won != 1 {
		t.Errorf("лок одновременно взяли %d процессов, want 1", won)
	}
}

// swapHook имитирует худший момент: сразу после того как Unlock прочитал
// значение лока, лок истекает и его берёт процесс B.
type swapHook struct {
	key, tokB string
	other     *redis.Client
	fired     atomic.Bool
}

func (h *swapHook) DialHook(next redis.DialHook) redis.DialHook { return next }

func (h *swapHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func (h *swapHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		err := next(ctx, cmd)
		args := cmd.Args()
		if cmd.Name() == "get" && len(args) > 1 && args[1] == h.key && h.fired.CompareAndSwap(false, true) {
			h.other.Set(ctx, h.key, h.tokB, time.Minute)
		}
		return err
	}
}

// Проверка и удаление должны быть одной атомарной операцией. Если Unlock
// читает лок отдельной командой GET, хук между GET и следующей командой
// отдаёт лок процессу B: неатомарный Unlock удалит чужой лок.
func TestUnlockAtomic(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	ctx := testCtx(t)
	key := prefix + "lock:atomic"
	tokA, ok, err := TryLock(ctx, rdb, key, time.Minute)
	if err != nil || !ok {
		t.Fatalf("TryLock = (%v, %v)", ok, err)
	}
	other := redis.NewClient(rdb.Options())
	t.Cleanup(func() { other.Close() })
	hook := &swapHook{key: key, tokB: "токен-B", other: other}
	rdb.AddHook(hook)

	got, err := Unlock(ctx, rdb, key, tokA)
	if err != nil {
		t.Fatalf("Unlock = %v", err)
	}
	val, _ := other.Get(ctx, key).Result()
	if hook.fired.Load() {
		// Unlock читал лок отдельным GET, и в этот момент лок перешёл к B
		if val != "токен-B" {
			t.Fatalf("Unlock(A) снял лок процесса B, взятый между GET и DEL: проверку и удаление делай одним Lua-скриптом")
		}
		return
	}
	if !got || val != "" {
		t.Fatalf("Unlock своим токеном = %v, ключ = %q, want true и ключ удалён", got, val)
	}
}
