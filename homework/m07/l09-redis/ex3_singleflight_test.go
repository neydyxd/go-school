package l09redis

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"

	"goschool/homework/internal/testredis"
)

// getCounter это хук go-redis: считает выполненные команды GET.
type getCounter struct {
	n       atomic.Int64
	target  int64
	reached chan struct{}
	once    sync.Once
}

func (h *getCounter) DialHook(next redis.DialHook) redis.DialHook { return next }

func (h *getCounter) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func (h *getCounter) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		err := next(ctx, cmd)
		if cmd.Name() == "get" && h.n.Add(1) >= h.target {
			h.once.Do(func() { close(h.reached) })
		}
		return err
	}
}

// 100 запросов одновременно промахиваются мимо кеша. База «медленная»:
// отвечает только после того, как все 100 сделали GET и увидели промах.
// В базу должен пойти ровно один.
func TestGetOnceStampede(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	const n = 100
	hook := &getCounter{target: n, reached: make(chan struct{})}
	rdb.AddHook(hook)

	db := newFakeDB()
	db.release = make(chan struct{})
	c := NewProductCache(rdb, prefix, time.Minute, db.load)

	var wg sync.WaitGroup
	errs := make(chan error, n)
	for range n {
		wg.Go(func() {
			p, err := c.GetOnce(testCtx(t), 42)
			if err == nil && p.Name != "Кружка" {
				t.Errorf("GetOnce(42) = %+v", p)
			}
			errs <- err
		})
	}
	select {
	case <-hook.reached:
	case <-time.After(5 * time.Second):
		close(db.release)
		wg.Wait()
		t.Fatalf("за 5 с выполнено только %d GET из %d: каждый вызов GetOnce сначала читает кеш", hook.n.Load(), n)
	}
	close(db.release) // база «ответила»
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("GetOnce(42) = %v", err)
		}
	}
	if got := db.calls.Load(); got != 1 {
		t.Errorf("при %d одновременных промахах база вызвана %d раз, want 1", n, got)
	}
	if got, err := c.GetOnce(testCtx(t), 42); err != nil || got.Price != 59900 {
		t.Errorf("GetOnce(42) из кеша = (%+v, %v)", got, err)
	}
	if got := db.calls.Load(); got != 1 {
		t.Errorf("после прогрева кеша база вызвана снова (%d раз)", got)
	}
}

func TestGetOnceError(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	db := newFakeDB()
	c := NewProductCache(rdb, prefix, time.Minute, db.load)
	if _, err := c.GetOnce(testCtx(t), 404); !errors.Is(err, ErrNotFound) {
		t.Errorf("GetOnce(404) = %v, want ErrNotFound", err)
	}
}

type tagKey struct{}

// lateHook задерживает «опоздавшую» горутину (помечена значением в ctx): после
// её первого GET (промах) она ждёт, пока ведущий вызов GetOnce полностью
// завершится. В Do она входит, когда полёт уже окончен и значение лежит в Redis.
type lateHook struct {
	first      atomic.Bool
	missed     chan struct{} // закрывается после первого GET опоздавшей
	leaderDone chan struct{}
}

func (h *lateHook) DialHook(next redis.DialHook) redis.DialHook { return next }

func (h *lateHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func (h *lateHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		err := next(ctx, cmd)
		if cmd.Name() == "get" && ctx.Value(tagKey{}) != nil && h.first.CompareAndSwap(false, true) {
			close(h.missed)
			<-h.leaderDone
		}
		return err
	}
}

// Опоздавший: промахнулся, пока ведущий полёт шёл, а вошёл в singleflight уже
// после его завершения. Без повторной проверки кеша внутри Do он начнёт новый
// полёт и снова сходит в базу. Сценарий детерминирован: шаги синхронизированы
// каналами и хуком go-redis, без sleep.
func TestGetOnceLateComer(t *testing.T) {
	t.Parallel()
	rdb, prefix := testredis.New(t)
	hook := &lateHook{missed: make(chan struct{}), leaderDone: make(chan struct{})}
	rdb.AddHook(hook)

	db := newFakeDB()
	db.release = make(chan struct{})
	loading := make(chan struct{})
	var once sync.Once
	db.onLoad = func() { once.Do(func() { close(loading) }) }
	c := NewProductCache(rdb, prefix, time.Minute, db.load)

	waitFor := func(ch <-chan struct{}, what string) {
		t.Helper()
		select {
		case <-ch:
		case <-time.After(5 * time.Second):
			close(db.release)
			t.Fatalf("за 5 с не дождались: %s", what)
		}
	}

	// 1. ведущий промахивается и висит в загрузке
	var leaderErr error
	go func() {
		defer close(hook.leaderDone)
		_, leaderErr = c.GetOnce(testCtx(t), 42)
	}()
	waitFor(loading, "ведущий GetOnce пошёл в базу")

	// 2. опоздавший делает GET (промах) и ждёт в хуке конца ведущего
	late := start(func() error {
		_, err := c.GetOnce(context.WithValue(testCtx(t), tagKey{}, true), 42)
		return err
	})
	waitFor(hook.missed, "опоздавший GetOnce сначала читает кеш (шаг 1 условия)")

	// 3. база отвечает, ведущий кладёт значение в кеш и завершается
	close(db.release)
	if err := late.wait(t); err != nil {
		t.Fatalf("GetOnce опоздавшего = %v", err)
	}
	if leaderErr != nil {
		t.Fatalf("GetOnce ведущего = %v", leaderErr)
	}
	if n := db.calls.Load(); n != 1 {
		t.Errorf("база вызвана %d раз, want 1: внутри group.Do сначала перепроверь кеш, полёт мог закончиться, пока ты ждал", n)
	}
}

type lateCall struct {
	done chan struct{}
	err  error
}

func start(f func() error) *lateCall {
	c := &lateCall{done: make(chan struct{})}
	go func() {
		defer close(c.done)
		c.err = f()
	}()
	return c
}

func (c *lateCall) wait(t *testing.T) error {
	t.Helper()
	select {
	case <-c.done:
		return c.err
	case <-time.After(10 * time.Second):
		t.Fatalf("GetOnce не завершился за 10 с")
		return nil
	}
}
