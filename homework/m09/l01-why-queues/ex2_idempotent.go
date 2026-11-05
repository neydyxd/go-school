package l01whyqueues

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNoAccount: у пользователя нет счёта.
var ErrNoAccount = errors.New("счёт не найден")

// ApplyPayment зачисляет amount на счёт userID (таблица accounts) ровно один раз
// для каждого paymentID, сколько бы раз брокер ни доставил одно и то же сообщение
// и сколько бы копий ни пришло одновременно.
//
// Всё в ОДНОЙ транзакции:
//  1. INSERT INTO processed_payments (payment_id, user_id, amount) ...
//     ON CONFLICT (payment_id) DO NOTHING. Вставилось 0 строк: платёж уже
//     обработан, верни (false, nil) и ничего не меняй.
//  2. UPDATE accounts SET balance = balance + amount WHERE user_id = ...
//     Счёта нет (0 строк): ErrNoAccount, и отметка из шага 1 тоже должна
//     откатиться, чтобы платёж можно было применить, когда счёт появится.
//
// Успех: (true, nil). Почему не «SELECT, есть ли отметка, потом INSERT»: две
// параллельные копии обе увидят «нет отметки». Уникальный ключ в базе решает
// гонку за тебя.
func ApplyPayment(ctx context.Context, pool *pgxpool.Pool, paymentID string, userID, amount int64) (bool, error) {
	panic("TODO: ApplyPayment")
}
