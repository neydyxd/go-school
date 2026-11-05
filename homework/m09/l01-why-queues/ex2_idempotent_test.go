package l01whyqueues

import (
	"errors"
	"reflect"
	"sync"
	"testing"

	"goschool/homework/internal/testdb"
)

func TestApplyPaymentOnce(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)

	applied, err := ApplyPayment(ctx, pool, "pay-1", 1, 50)
	if err != nil || !applied {
		t.Fatalf(`ApplyPayment("pay-1") первый раз = (%v, %v), want (true, nil)`, applied, err)
	}
	applied, err = ApplyPayment(ctx, pool, "pay-1", 1, 50)
	if err != nil || applied {
		t.Fatalf(`ApplyPayment("pay-1") повторно = (%v, %v), want (false, nil)`, applied, err)
	}
	applied, err = ApplyPayment(ctx, pool, "pay-2", 1, 7)
	if err != nil || !applied {
		t.Fatalf(`ApplyPayment("pay-2") = (%v, %v), want (true, nil)`, applied, err)
	}
	got := testdb.Rows(t, pool, `SELECT user_id, balance FROM accounts ORDER BY user_id`)
	want := [][]string{{"1", "157"}, {"2", "0"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("балансы = %v, want %v (100 + 50 один раз + 7)", got, want)
	}
}

func TestApplyPaymentConcurrentDuplicates(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)

	const copies = 8
	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		applied int
		errs    []error
	)
	start := make(chan struct{})
	for range copies {
		wg.Go(func() {
			<-start
			ok, err := ApplyPayment(ctx, pool, "pay-dup", 2, 30)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				errs = append(errs, err)
			}
			if ok {
				applied++
			}
		})
	}
	close(start)
	wg.Wait()

	if len(errs) > 0 {
		t.Fatalf("%d параллельных копий одного сообщения: ошибки %v, want ни одной (дубль это не ошибка, а (false, nil))", copies, errs)
	}
	if applied != 1 {
		t.Errorf("%d параллельных копий: применено %d раз, want 1", copies, applied)
	}
	if got := testdb.Rows(t, pool, `SELECT balance FROM accounts WHERE user_id = 2`); got[0][0] != "30" {
		t.Errorf("баланс пользователя 2 = %s, want 30", got[0][0])
	}
}

func TestApplyPaymentNoAccountRollsBack(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)

	applied, err := ApplyPayment(ctx, pool, "pay-9", 9, 10)
	if !errors.Is(err, ErrNoAccount) || applied {
		t.Fatalf(`ApplyPayment для несуществующего счёта = (%v, %v), want (false, ErrNoAccount)`, applied, err)
	}
	if got := testdb.Rows(t, pool, `SELECT count(*) FROM processed_payments`); got[0][0] != "0" {
		t.Fatalf("после ErrNoAccount в processed_payments %s строк, want 0: отметка должна откатиться вместе с зачислением", got[0][0])
	}

	// счёт появился: тот же платёж теперь применяется
	testdb.Exec(t, pool, `INSERT INTO accounts (user_id, balance) VALUES (9, 0)`)
	applied, err = ApplyPayment(ctx, pool, "pay-9", 9, 10)
	if err != nil || !applied {
		t.Fatalf(`ApplyPayment после создания счёта = (%v, %v), want (true, nil)`, applied, err)
	}
	if got := testdb.Rows(t, pool, `SELECT balance FROM accounts WHERE user_id = 9`); got[0][0] != "10" {
		t.Errorf("баланс пользователя 9 = %s, want 10", got[0][0])
	}
}
