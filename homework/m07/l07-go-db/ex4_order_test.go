package l07godb

import (
	"errors"
	"reflect"
	"testing"

	"goschool/homework/internal/testdb"
)

func TestCreateOrder(t *testing.T) {
	t.Parallel()
	repo, pool := newRepo(t)
	ctx := testCtx(t)

	id, err := repo.CreateOrder(ctx, 1, []Item{{"tea", 2, 15000}, {"cup", 1, 49900}})
	if err != nil || id != 1 {
		t.Fatalf("CreateOrder(1, tea x2 + cup) = (%d, %v), want (1, nil)", id, err)
	}
	got := testdb.Rows(t, pool, `SELECT o.id, o.user_id, o.total, i.sku, i.qty, i.price
		FROM orders o JOIN order_items i ON i.order_id = o.id ORDER BY i.sku`)
	want := [][]string{{"1", "1", "79900", "cup", "1", "49900"}, {"1", "1", "79900", "tea", "2", "15000"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("заказ в базе = %v, want %v", got, want)
	}

	fails := []struct {
		name    string
		user    int64
		items   []Item
		wantErr error
	}{
		{"пустой заказ", 1, nil, ErrEmptyOrder},
		{"нет пользователя", 99, []Item{{"tea", 1, 100}}, ErrNotFound},
		{"вторая позиция нарушает CHECK", 2, []Item{{"tea", 1, 100}, {"broken", 0, 100}}, nil},
	}
	for _, f := range fails {
		_, err := repo.CreateOrder(ctx, f.user, f.items)
		if err == nil {
			t.Fatalf("%s: CreateOrder(%d, %v) = nil, want ошибку", f.name, f.user, f.items)
		}
		if f.wantErr != nil && !errors.Is(err, f.wantErr) {
			t.Fatalf("%s: CreateOrder(%d, %v) = %v, want %v", f.name, f.user, f.items, err, f.wantErr)
		}
	}
	// после трёх неудач в базе только первый заказ: ни «сирот», ни половины позиций
	counts := testdb.Rows(t, pool, `SELECT (SELECT count(*) FROM orders), (SELECT count(*) FROM order_items)`)
	if !reflect.DeepEqual(counts, [][]string{{"1", "2"}}) {
		t.Errorf("заказов и позиций = %v, want [[1 2]]: неудачный заказ оставил следы, нужна одна транзакция", counts)
	}
	// соединения вернулись в пул: если транзакция не закрыта, этот запрос повиснет
	if s := pool.Stat(); s.AcquiredConns() != 0 {
		t.Errorf("после CreateOrder занято %d соединений, want 0: транзакцию нужно завершать Commit или Rollback", s.AcquiredConns())
	}
}
