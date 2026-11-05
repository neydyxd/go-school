package l01whyqueues

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"goschool/homework/internal/testdb"
)

// countingDB считает обращения к базе. Claim обязан забирать задачу ОДНИМ
// оператором: если «найти» и «пометить» это два запроса через пул, блокировка
// первого отпускается до второго, и два воркера могут взять одну задачу.
type countingDB struct {
	DBTX
	calls int
}

func (c *countingDB) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	c.calls++
	return c.DBTX.Exec(ctx, sql, args...)
}

func (c *countingDB) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	c.calls++
	return c.DBTX.QueryRow(ctx, sql, args...)
}

func TestClaimSingleStatement(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)
	_, _ = Enqueue(ctx, pool, "k", "p", time.Now().Add(-time.Hour))

	for _, tt := range []struct {
		name   string
		wantOK bool
	}{{"есть задача", true}, {"очередь пуста", false}} {
		db := &countingDB{DBTX: pool}
		_, ok, err := Claim(ctx, db, "w")
		if err != nil || ok != tt.wantOK {
			t.Fatalf("%s: Claim = (_, %v, %v), want (_, %v, nil)", tt.name, ok, err, tt.wantOK)
		}
		if db.calls != 1 {
			t.Errorf("%s: Claim сделал %d запросов к базе, want 1: найти и пометить задачу нужно одним UPDATE ... WHERE id = (SELECT ... FOR UPDATE SKIP LOCKED LIMIT 1)", tt.name, db.calls)
		}
	}
}

func TestClaimBasics(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)

	if j, ok, err := Claim(ctx, pool, "w1"); err != nil || ok {
		t.Fatalf("Claim на пустой очереди = (%+v, %v, %v), want (Job{}, false, nil)", j, ok, err)
	}

	now := time.Now()
	idLate, _ := Enqueue(ctx, pool, "email", "поздняя", now.Add(-time.Minute))
	idEarly, _ := Enqueue(ctx, pool, "email", "ранняя", now.Add(-time.Hour))
	idSame, _ := Enqueue(ctx, pool, "email", "такая же поздняя, но id больше", now.Add(-time.Minute))
	_, _ = Enqueue(ctx, pool, "email", "будущая", now.Add(time.Hour))

	var got []int64
	for range 3 {
		j, ok, err := Claim(ctx, pool, "w1")
		if err != nil || !ok {
			t.Fatalf("Claim = (%+v, %v, %v), want задачу", j, ok, err)
		}
		got = append(got, j.ID)
		if j.Attempts != 1 || j.Kind != "email" {
			t.Errorf("Claim вернула %+v, want Kind=email и Attempts=1", j)
		}
	}
	if want := []int64{idEarly, idLate, idSame}; !reflect.DeepEqual(got, want) {
		t.Errorf("порядок Claim = %v, want %v (по run_at, при равенстве по id)", got, want)
	}
	if j, ok, err := Claim(ctx, pool, "w1"); err != nil || ok {
		t.Errorf("Claim, когда осталась только задача из будущего = (%+v, %v, %v), want (Job{}, false, nil)", j, ok, err)
	}
	rows := testdb.Rows(t, pool, `SELECT state, coalesce(locked_by, 'NULL') FROM queue_jobs WHERE id = $1`, idEarly)
	if !reflect.DeepEqual(rows, [][]string{{"running", "w1"}}) {
		t.Errorf("забранная задача в базе: %v, want [[running w1]]", rows)
	}
}

// Две транзакции забирают задачи одновременно. Первая держит блокировку, вторая
// обязана взять ДРУГУЮ задачу и не ждать первую. Без sleep: первая транзакция
// просто не коммитится, пока вторая работает.
func TestClaimSkipLocked(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)
	for i := range 3 {
		_, _ = Enqueue(ctx, pool, "k", fmt.Sprint(i), time.Now().Add(-time.Hour))
	}

	tx1, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx1.Rollback(ctx)
	j1, ok, err := Claim(ctx, tx1, "w1")
	if err != nil || !ok {
		t.Fatalf("Claim в первой транзакции = (%+v, %v, %v)", j1, ok, err)
	}

	tx2, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx2.Rollback(ctx)
	ctx2, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	j2, ok, err := Claim(ctx2, tx2, "w2")
	if errors.Is(ctx2.Err(), context.DeadlineExceeded) {
		t.Fatal("второй Claim ждал блокировку первой транзакции 3 секунды: нужен FOR UPDATE SKIP LOCKED")
	}
	if err != nil || !ok {
		t.Fatalf("Claim во второй транзакции = (%+v, %v, %v), want свободную задачу", j2, ok, err)
	}
	if j2.ID == j1.ID {
		t.Fatalf("две транзакции получили одну задачу %d", j1.ID)
	}

	// Вторая транзакция откатилась (воркер упал до коммита): её задача снова доступна.
	if err := tx2.Rollback(ctx); err != nil {
		t.Fatal(err)
	}
	if err := tx1.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	j3, ok, err := Claim(ctx, pool, "w3")
	if err != nil || !ok || j3.ID != j2.ID {
		t.Fatalf("после отката второй транзакции Claim = (%+v, %v, %v), want снова задачу %d", j3, ok, err, j2.ID)
	}
}

func TestClaimConcurrentWorkers(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)
	const jobs, workers = 40, 4
	for i := range jobs {
		_, _ = Enqueue(ctx, pool, "k", fmt.Sprint(i), time.Now().Add(-time.Hour))
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		seen = map[int64]int{}
		errs []error
	)
	for w := range workers {
		wg.Go(func() {
			for {
				j, ok, err := Claim(ctx, pool, fmt.Sprintf("w%d", w))
				if err == nil && ok {
					err = Complete(ctx, pool, j.ID)
				}
				mu.Lock()
				if err != nil {
					errs = append(errs, err)
				}
				if ok {
					seen[j.ID]++
				}
				mu.Unlock()
				if err != nil || !ok {
					return
				}
			}
		})
	}
	wg.Wait()
	if len(errs) > 0 {
		t.Fatalf("ошибки воркеров: %v", errs)
	}
	for id, n := range seen {
		if n != 1 {
			t.Errorf("задача %d обработана %d раз, want 1", id, n)
		}
	}
	if len(seen) != jobs {
		t.Errorf("обработано %d задач из %d", len(seen), jobs)
	}
	if got := testdb.Rows(t, pool, `SELECT count(*) FROM queue_jobs WHERE state <> 'done'`); got[0][0] != "0" {
		t.Errorf("невыполненных задач осталось %s, want 0", got[0][0])
	}
}

func TestCompleteAndFail(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)
	id, _ := Enqueue(ctx, pool, "k", "p", time.Now().Add(-time.Hour))

	if err := Complete(ctx, pool, id); err == nil {
		t.Error("Complete задачи в состоянии available = nil, want ошибку")
	}
	if err := Fail(ctx, pool, id, "x", time.Now(), 3); err == nil {
		t.Error("Fail задачи в состоянии available = nil, want ошибку")
	}
	if err := Complete(ctx, pool, 999); err == nil {
		t.Error("Complete несуществующей задачи = nil, want ошибку")
	}

	retryAt := time.Now().Add(-time.Second).UTC().Truncate(time.Second)
	for attempt := 1; attempt <= 3; attempt++ {
		j, ok, err := Claim(ctx, pool, "w")
		if err != nil || !ok || j.ID != id || j.Attempts != attempt {
			t.Fatalf("попытка %d: Claim = (%+v, %v, %v), want задачу %d с Attempts=%d", attempt, j, ok, err, id, attempt)
		}
		if err := Fail(ctx, pool, id, fmt.Sprintf("ошибка %d", attempt), retryAt, 3); err != nil {
			t.Fatalf("попытка %d: Fail = %v", attempt, err)
		}
		row := testdb.Rows(t, pool, `SELECT state, coalesce(locked_by, 'NULL'), last_error, run_at FROM queue_jobs WHERE id = $1`, id)[0]
		wantState := "available"
		if attempt == 3 {
			wantState = "dead"
		}
		want := []string{wantState, "NULL", fmt.Sprintf("ошибка %d", attempt), retryAt.Format(time.RFC3339)}
		if !reflect.DeepEqual(row, want) {
			t.Errorf("после Fail на попытке %d из 3: строка %v, want %v", attempt, row, want)
		}
	}
	if j, ok, _ := Claim(ctx, pool, "w"); ok {
		t.Errorf("мёртвую задачу снова забрали: %+v", j)
	}
}
