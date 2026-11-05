package l06transactions

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrInvalidAmount     = errors.New("сумма перевода должна быть положительной")
	ErrSameAccount       = errors.New("перевод на тот же счёт")
	ErrAccountNotFound   = errors.New("счёт не найден")
	ErrInsufficientFunds = errors.New("недостаточно средств")
)

// Transfer переводит amount копеек со счёта from на счёт to (таблица accounts)
// в одной транзакции.
//
// Ошибки (проверяются через errors.Is), при любой ошибке балансы не меняются:
//   - amount <= 0: ErrInvalidAmount;
//   - from == to: ErrSameAccount;
//   - одного из счетов нет: ErrAccountNotFound;
//   - на from меньше amount: ErrInsufficientFunds.
//
// Требования к конкурентности (их проверяют тесты):
//   - нет потерянного обновления: если параллельно кто-то меняет баланс,
//     его изменение не должно затереться;
//   - нет взаимоблокировок: переводы 1->2 и 2->1 одновременно не должны
//     падать с deadlock detected. Подсказка: блокируй оба счёта
//     в одном и том же порядке (по id), независимо от направления перевода.
func Transfer(ctx context.Context, pool *pgxpool.Pool, from, to, amount int64) error {
	panic("TODO: Transfer")
}
