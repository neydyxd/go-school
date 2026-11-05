package l03riveradvanced

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivertest"

	"goschool/homework/internal/testdb"
)

// fakeGateway ведёт себя как настоящий шлюз с ключами идемпотентности:
// списывает один раз на ключ. lostReply раз подряд «списывает, но ответ теряется».
type fakeGateway struct {
	mu        sync.Mutex
	keys      []string         // ключи всех вызовов
	charged   map[string]int64 // ключ -> списанная сумма
	lostReply int
	arrived   chan struct{} // если не nil, каждый вызов сообщает о себе
	barrier   chan struct{} // если не nil, вызов ждёт, пока его не закроют
}

var errTimeout = errors.New("шлюз: таймаут ответа")

func (g *fakeGateway) Charge(ctx context.Context, key string, _, amount int64) error {
	g.mu.Lock()
	g.keys = append(g.keys, key)
	if g.charged == nil {
		g.charged = map[string]int64{}
	}
	if _, ok := g.charged[key]; !ok {
		g.charged[key] = amount
	}
	lost := g.lostReply > 0
	if lost {
		g.lostReply--
	}
	barrier, arrived := g.barrier, g.arrived
	g.mu.Unlock()
	if arrived != nil {
		arrived <- struct{}{}
	}
	if barrier != nil {
		select {
		case <-barrier:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if lost {
		return errTimeout
	}
	return nil
}

func (g *fakeGateway) calls() (keys []string, charges int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return append([]string(nil), g.keys...), len(g.charged)
}

func TestChargeWorkerRedelivery(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	gw := &fakeGateway{}
	tw := rivertest.NewWorker(t, riverpgxv5.New(nil), &river.Config{}, &ChargeWorker{Pool: pool, Gateway: gw})
	args := ChargeArgs{PaymentID: "p-1", UserID: 1, Amount: 500}

	for run := 1; run <= 2; run++ {
		res, err := tw.Work(ctx, t, beginTx(t, pool), args, nil)
		if err != nil || res.EventKind != river.EventKindJobCompleted {
			t.Fatalf("доставка %d: Work = (%v, %v), want (job_completed, nil)", run, res.EventKind, err)
		}
	}
	if keys, _ := gw.calls(); len(keys) != 1 {
		t.Errorf("шлюз вызван %d раз на две доставки, want 1: проверь payments до вызова шлюза", len(keys))
	}
	if got := testdb.Rows(t, pool, `SELECT payment_id, amount FROM payments`); len(got) != 1 || got[0][1] != "500" {
		t.Errorf("payments = %v, want одну строку p-1 на 500", got)
	}
}

func TestChargeWorkerLostReply(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	gw := &fakeGateway{lostReply: 1} // деньги списаны, но ответ потерялся
	tw := rivertest.NewWorker(t, riverpgxv5.New(nil), &river.Config{}, &ChargeWorker{Pool: pool, Gateway: gw})
	args := ChargeArgs{PaymentID: "p-2", UserID: 1, Amount: 300}

	res, err := tw.Work(ctx, t, beginTx(t, pool), args, nil)
	if err == nil || res.EventKind != river.EventKindJobFailed {
		t.Fatalf("потерянный ответ: Work = (%v, %v), want (job_failed, ошибку)", res.EventKind, err)
	}
	if n := testdb.Rows(t, pool, `SELECT count(*) FROM payments`)[0][0]; n != "0" {
		t.Fatalf("после ошибки шлюза в payments %s строк, want 0", n)
	}
	// повтор: это новая попытка (а в тесте даже новая задача с другим id)
	res, err = tw.Work(ctx, t, beginTx(t, pool), args, nil)
	if err != nil || res.EventKind != river.EventKindJobCompleted {
		t.Fatalf("повтор: Work = (%v, %v), want (job_completed, nil)", res.EventKind, err)
	}
	keys, charges := gw.calls()
	if charges != 1 {
		t.Errorf("шлюз списал %d раз (ключи %v), want 1: ключ идемпотентности должен зависеть только от PaymentID", charges, keys)
	}
	for _, k := range keys {
		if k != IdempotencyKey("p-2") {
			t.Errorf("ключ %q, want %q", k, IdempotencyKey("p-2"))
		}
	}
	if n := testdb.Rows(t, pool, `SELECT count(*) FROM payments`)[0][0]; n != "1" {
		t.Errorf("в payments %s строк, want 1", n)
	}
}

// goTB даёт вызывать rivertest из обычной горутины: t.FailNow вне тестовой
// горутины запрещён, поэтому Fatal здесь превращается в Error и выход только
// из этой горутины (runtime.Goexit, отложенные вызовы выполняются).
type goTB struct{ testing.TB }

func (g goTB) Fatal(args ...any)                 { g.TB.Error(args...); runtime.Goexit() }
func (g goTB) Fatalf(format string, args ...any) { g.TB.Errorf(format, args...); runtime.Goexit() }
func (g goTB) FailNow()                          { g.TB.Fail(); runtime.Goexit() }

// Две копии одной задачи выполняются одновременно: обе прошли проверку payments,
// обе дошли до шлюза. Вторая вставка в payments не должна падать.
// Копии работают в обычных горутинах (не зависят от -parallel), ошибки
// проверяются в главной горутине теста.
func TestChargeWorkerConcurrentCopies(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	gw := &fakeGateway{arrived: make(chan struct{}, 2), barrier: make(chan struct{})}
	args := ChargeArgs{PaymentID: "p-3", UserID: 2, Amount: 100}

	var wg sync.WaitGroup
	errs := make([]error, 2)
	for i := range 2 {
		tw := rivertest.NewWorker(t, riverpgxv5.New(nil), &river.Config{}, &ChargeWorker{Pool: pool, Gateway: gw})
		tx := beginTx(t, pool)
		wg.Go(func() {
			_, errs[i] = tw.Work(ctx, goTB{t}, tx, args, nil)
		})
	}
	// ждём, пока обе копии дойдут до шлюза, и отпускаем их вместе
	timeout := time.After(10 * time.Second)
	for i := range 2 {
		select {
		case <-gw.arrived:
		case <-timeout:
			close(gw.barrier)
			wg.Wait()
			t.Fatalf("за 10 секунд до шлюза дошло %d копий из 2: проверка payments и вызов шлюза не должны быть под общей блокировкой", i)
		}
	}
	close(gw.barrier)
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Errorf("копия %d: Work = %v, want nil: дубль в payments это не ошибка (ON CONFLICT DO NOTHING)", i+1, err)
		}
	}
	if n := testdb.Rows(t, pool, `SELECT count(*) FROM payments`)[0][0]; n != "1" {
		t.Errorf("в payments %s строк, want 1", n)
	}
	if _, charges := gw.calls(); charges != 1 {
		t.Errorf("шлюз списал %d раз, want 1", charges)
	}
}
