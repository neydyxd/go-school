package l06transactions

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if from == to {
		return ErrSameAccount
	}
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) // после Commit ничего не делает

	// ORDER BY id + FOR UPDATE: оба перевода (1->2 и 2->1) сначала ждут
	// строку 1, поэтому друг друга по кругу не ждут. Балансы читаем уже
	// под блокировкой, значит чужое изменение не потеряется.
	rows, _ := tx.Query(ctx, `SELECT id, balance FROM accounts WHERE id IN ($1, $2) ORDER BY id FOR UPDATE`, from, to)
	balance := map[int64]int64{}
	var id, b int64
	if _, err := pgx.ForEachRow(rows, []any{&id, &b}, func() error {
		balance[id] = b
		return nil
	}); err != nil {
		return err
	}
	if len(balance) != 2 {
		return ErrAccountNotFound
	}
	if balance[from] < amount {
		return ErrInsufficientFunds
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = balance - $2 WHERE id = $1`, from, amount); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE accounts SET balance = balance + $2 WHERE id = $1`, to, amount); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
