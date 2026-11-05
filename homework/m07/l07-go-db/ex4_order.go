package l07godb

import (
	"context"
	"errors"
)

var ErrEmptyOrder = errors.New("в заказе нет позиций")

// Item это позиция заказа: количество и цена за штуку в копейках.
type Item struct {
	SKU   string
	Qty   int
	Price int64
}

// CreateOrder создаёт заказ пользователя userID с позициями items и возвращает id заказа.
// total заказа это сумма Qty*Price по всем позициям.
//
// Заказ и позиции пишутся в ОДНОЙ транзакции: если любая вставка упала
// (например, Qty <= 0 нарушает CHECK), в базе не должно остаться ни заказа,
// ни части позиций. Паттерн: Begin, defer tx.Rollback(ctx), работа, tx.Commit(ctx).
//
//   - items пустой: ErrEmptyOrder, в базу не ходить;
//   - пользователя нет (нарушение внешнего ключа, код 23503): ErrNotFound;
//   - прочие ошибки базы возвращай обёрнутыми через %w.
func (r *Repo) CreateOrder(ctx context.Context, userID int64, items []Item) (int64, error) {
	panic("TODO: CreateOrder")
}
