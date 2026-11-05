package l07godb

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
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
	if len(items) == 0 {
		return 0, ErrEmptyOrder
	}
	var total int64
	for _, it := range items {
		total += int64(it.Qty) * it.Price
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin: %w", err)
	}
	// Откат на любом раннем return. После успешного Commit он ничего не делает.
	defer tx.Rollback(ctx)

	var orderID int64
	err = tx.QueryRow(ctx, `INSERT INTO orders (user_id, total) VALUES ($1, $2) RETURNING id`, userID, total).Scan(&orderID)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("insert order: %w", err)
	}
	for _, it := range items {
		if _, err := tx.Exec(ctx,
			`INSERT INTO order_items (order_id, sku, qty, price) VALUES ($1, $2, $3, $4)`,
			orderID, it.SKU, it.Qty, it.Price); err != nil {
			return 0, fmt.Errorf("insert item %s: %w", it.SKU, err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit: %w", err)
	}
	return orderID, nil
}
