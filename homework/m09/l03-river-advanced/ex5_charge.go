package l03riveradvanced

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// ChargeArgs: задача «списать деньги за заказ».
type ChargeArgs struct {
	PaymentID string `json:"payment_id"`
	UserID    int64  `json:"user_id"`
	Amount    int64  `json:"amount"`
}

func (ChargeArgs) Kind() string { return "charge" }

// Gateway это платёжный шлюз. Как у настоящих провайдеров (Stripe, ЮKassa),
// он принимает ключ идемпотентности: второй запрос с тем же ключом не списывает
// деньги повторно, а возвращает результат первого.
type Gateway interface {
	Charge(ctx context.Context, idempotencyKey string, userID, amount int64) error
}

// ChargeWorker списывает деньги и записывает платёж в таблицу payments.
type ChargeWorker struct {
	river.WorkerDefaults[ChargeArgs]

	Pool    *pgxpool.Pool
	Gateway Gateway
}

// IdempotencyKey это ключ для шлюза: "charge:" + PaymentID. Один и тот же для
// всех попыток и всех копий задачи с этим платежом.
func IdempotencyKey(paymentID string) string { return "charge:" + paymentID }

// Work:
//  1. Платёж уже записан (SELECT EXISTS ... FROM payments WHERE payment_id = ...):
//     верни nil, шлюз не вызывай.
//  2. Gateway.Charge с ключом IdempotencyKey(PaymentID). Ошибка: верни её.
//  3. INSERT INTO payments (payment_id, user_id, amount) ...
//     ON CONFLICT (payment_id) DO NOTHING.
//
// Почему нужны и шаг 1, и ключ: шаг 1 экономит вызов шлюза при повторной
// доставке, а ключ спасает, когда процесс упал между шагами 2 и 3 или шлюз
// списал деньги, но ответ потерялся по таймауту. ON CONFLICT спасает от двух
// копий задачи, которые выполняются одновременно.
//
// Блокировок (advisory lock по payment_id и т.п.) не бери: тест запускает две
// копии одновременно и ждёт, что обе дойдут до шлюза и обе завершатся без
// ошибки. Защиту от двойного списания здесь дают ключ шлюза и ON CONFLICT.
func (w *ChargeWorker) Work(ctx context.Context, job *river.Job[ChargeArgs]) error {
	panic("TODO: Work")
}
