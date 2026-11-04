package l07problems

import (
	"errors"
	"sync"
)

var (
	ErrInvalidAmount = errors.New("сумма перевода должна быть положительной")
	ErrSameAccount   = errors.New("перевод самому себе")
	ErrInsufficient  = errors.New("недостаточно средств")
)

// Account: банковский счёт со своим мьютексом.
type Account struct {
	ID      int
	mu      sync.Mutex
	balance int
}

// NewAccount создаёт счёт с начальным балансом.
func NewAccount(id, balance int) *Account {
	return &Account{ID: id, balance: balance}
}

// Balance возвращает текущий баланс (безопасно из многих горутин).
func (a *Account) Balance() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance
}

// Transfer переводит amount со счёта from на счёт to атомарно: под
// мьютексами ОБОИХ счетов, чтобы никто не увидел деньги «в полёте».
//   - amount <= 0: ErrInvalidAmount.
//   - from == to (один и тот же счёт): ErrSameAccount, и без дедлока.
//   - денег не хватает: ErrInsufficient, балансы не меняются.
//
// Главное: параллельные переводы A→B и B→A не должны приводить к дедлоку.
// Правило: захватывай мьютексы всегда в одном порядке, по возрастанию ID
// (тест это проверяет).
func Transfer(from, to *Account, amount int) error {
	panic("TODO: реализуй Transfer")
}
