package l01whyqueues

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"goschool/homework/internal/testdb"
)

func TestPlaceOrder(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)

	id, err := PlaceOrder(ctx, pool, 1, 2)
	if err != nil || id == 0 {
		t.Fatalf("PlaceOrder(кофе, 2) = (%d, %v), want (id, nil)", id, err)
	}
	rows := testdb.Rows(t, pool, `SELECT topic, payload::text FROM outbox`)
	if len(rows) != 1 || rows[0][0] != "order_created" {
		t.Fatalf("outbox после заказа = %v, want одно событие order_created", rows)
	}
	var ev OrderCreated
	if err := json.Unmarshal([]byte(rows[0][1]), &ev); err != nil || ev != (OrderCreated{OrderID: id, ProductID: 1, Qty: 2}) {
		t.Errorf("payload = %s (%v), want {order_id:%d, product_id:1, qty:2}", rows[0][1], err, id)
	}
	if got := testdb.Rows(t, pool, `SELECT stock FROM products WHERE id = 1`); got[0][0] != "3" {
		t.Errorf("остаток кофе = %s, want 3", got[0][0])
	}
}

func TestPlaceOrderRollback(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)

	tests := []struct {
		name      string
		productID int64
		qty       int
		wantErr   error
	}{
		{"чая нет совсем", 2, 1, ErrOutOfStock},
		{"кофе меньше, чем просят", 1, 6, ErrOutOfStock},
		{"ноль штук", 1, 0, ErrBadQty},
		{"отрицательное количество", 1, -3, ErrBadQty},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := PlaceOrder(ctx, pool, tt.productID, tt.qty)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("PlaceOrder(%d, %d) error = %v, want %v", tt.productID, tt.qty, err, tt.wantErr)
			}
			got := testdb.Rows(t, pool, `SELECT (SELECT count(*) FROM orders), (SELECT count(*) FROM outbox),
				(SELECT string_agg(stock::text, ',' ORDER BY id) FROM products)`)
			if want := [][]string{{"0", "0", "5,0"}}; !reflect.DeepEqual(got, want) {
				t.Errorf("после ошибки (заказов, событий, остатки) = %v, want %v: всё должно откатиться", got, want)
			}
		})
	}
}

func TestRelayOnce(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)
	for i := 1; i <= 5; i++ {
		testdb.Exec(t, pool, fmt.Sprintf(`INSERT INTO outbox (topic, payload) VALUES ('t', '{"n": %d}')`, i))
	}
	unpublished := func() []string {
		var ids []string
		for _, r := range testdb.Rows(t, pool, `SELECT id FROM outbox WHERE published_at IS NULL ORDER BY id`) {
			ids = append(ids, r[0])
		}
		return ids
	}

	// брокер падает на третьем событии
	var sent []int64
	broken := errors.New("брокер недоступен")
	n, err := RelayOnce(ctx, pool, 10, func(_ context.Context, e Event) error {
		if e.ID == 3 {
			return broken
		}
		sent = append(sent, e.ID)
		return nil
	})
	if n != 2 || !errors.Is(err, broken) {
		t.Fatalf("RelayOnce с падением на событии 3 = (%d, %v), want (2, обёрнутую ошибку брокера)", n, err)
	}
	if !reflect.DeepEqual(sent, []int64{1, 2}) {
		t.Errorf("опубликованы %v, want [1 2]", sent)
	}
	if got := unpublished(); !reflect.DeepEqual(got, []string{"3", "4", "5"}) {
		t.Errorf("неопубликованные после сбоя = %v, want [3 4 5]", got)
	}

	// брокер ожил, limit 2
	sent = nil
	var payloads []string
	n, err = RelayOnce(ctx, pool, 2, func(_ context.Context, e Event) error {
		sent = append(sent, e.ID)
		payloads = append(payloads, string(e.Payload))
		return nil
	})
	if n != 2 || err != nil || !reflect.DeepEqual(sent, []int64{3, 4}) {
		t.Fatalf("RelayOnce(limit=2) = (%d, %v), опубликованы %v, want (2, nil), [3 4]", n, err, sent)
	}
	if payloads[0] != `{"n": 3}` {
		t.Errorf("Payload события 3 = %s, want {\"n\": 3}", payloads[0])
	}
	if got := unpublished(); !reflect.DeepEqual(got, []string{"5"}) {
		t.Errorf("неопубликованные = %v, want [5]", got)
	}
}

// Ретранслятор A держит события 1-2 (публикует первое), и в этот момент запускается
// ретранслятор B. B обязан взять только 3-5 и не ждать A. Обрати внимание: B
// публикует 3-5 раньше, чем A опубликует 2. Глобального порядка у нескольких
// ретрансляторов со SKIP LOCKED нет, это осознанная плата за параллельность.
func TestRelayOnceConcurrent(t *testing.T) {
	t.Parallel()
	pool := newPool(t)
	ctx := testCtx(t)
	for i := 1; i <= 5; i++ {
		testdb.Exec(t, pool, fmt.Sprintf(`INSERT INTO outbox (topic, payload) VALUES ('t', '{"n": %d}')`, i))
	}

	published := map[int64]int{}
	var bErr error
	var bN int
	nA, errA := RelayOnce(ctx, pool, 2, func(_ context.Context, e Event) error {
		published[e.ID]++
		if e.ID == 1 { // A ещё держит блокировки на 1 и 2: запускаем B
			ctxB, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			bN, bErr = RelayOnce(ctxB, pool, 10, func(_ context.Context, e Event) error {
				published[e.ID]++
				return nil
			})
			if errors.Is(ctxB.Err(), context.DeadlineExceeded) {
				bErr = errors.New("ретранслятор B ждал блокировки A 3 секунды: нужен FOR UPDATE SKIP LOCKED")
			}
		}
		return nil
	})
	if errA != nil || nA != 2 {
		t.Fatalf("ретранслятор A = (%d, %v), want (2, nil)", nA, errA)
	}
	if bErr != nil || bN != 3 {
		t.Fatalf("ретранслятор B = (%d, %v), want (3, nil)", bN, bErr)
	}
	for id := int64(1); id <= 5; id++ {
		if published[id] != 1 {
			t.Errorf("событие %d опубликовано %d раз, want 1", id, published[id])
		}
	}
}
