package l01whyqueues

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrBadQty     = errors.New("количество должно быть больше нуля")
	ErrOutOfStock = errors.New("товара нет в нужном количестве")
)

// OrderCreated это событие в outbox (поле payload, JSON).
type OrderCreated struct {
	OrderID   int64 `json:"order_id"`
	ProductID int64 `json:"product_id"`
	Qty       int   `json:"qty"`
}

// PlaceOrder оформляет заказ qty штук товара productID и возвращает id заказа.
//
// qty <= 0: ErrBadQty, в базу не ходить. Дальше строго в ОДНОЙ транзакции и
// строго в таком порядке:
//  1. INSERT INTO orders (product_id, qty) ... RETURNING id;
//  2. INSERT INTO outbox (topic, payload): topic "order_created", payload это
//     JSON OrderCreated (json.Marshal, передай как []byte или string);
//  3. UPDATE products SET stock = stock - qty WHERE id = ... AND stock >= qty.
//     0 строк: ErrOutOfStock, и ни заказа, ни события в базе остаться не должно.
//
// Событие пишется в ту же базу и ту же транзакцию, что и заказ. В брокер его
// отправит отдельный процесс (RelayOnce), поэтому «заказ есть, а события нет»
// и «событие ушло, а заказа нет» невозможны.
func PlaceOrder(ctx context.Context, pool *pgxpool.Pool, productID int64, qty int) (int64, error) {
	panic("TODO: PlaceOrder")
}

// Event это строка outbox.
type Event struct {
	ID      int64
	Topic   string
	Payload []byte
}

// RelayOnce делает один проход ретранслятора outbox: берёт до limit
// неопубликованных событий (published_at IS NULL) по возрастанию id и по очереди
// отдаёт их в publish. Каждое успешно опубликованное помечает published_at = now().
//
// Работает в транзакции, строки берёт через FOR UPDATE SKIP LOCKED: два
// ретранслятора, запущенные одновременно, не публикуют одно событие дважды
// и не ждут друг друга. Цена: глобального порядка между ретрансляторами нет
// (B может опубликовать 3-5, пока A ещё публикует 1-2). Порядок гарантирован
// только внутри одного прохода одного ретранслятора; как сохранить его целиком,
// разобрано в уроке.
//
// Первая ошибка publish останавливает проход: внутри прохода событие 3 не
// уходит раньше 2, поэтому это событие и все следующие из прохода остаются
// неопубликованными, а уже опубликованные остаются помеченными (COMMIT).
// Вернуть число опубликованных и ошибку publish, обёрнутую через %w.
func RelayOnce(ctx context.Context, pool *pgxpool.Pool, limit int, publish func(ctx context.Context, e Event) error) (int, error) {
	panic("TODO: RelayOnce")
}
