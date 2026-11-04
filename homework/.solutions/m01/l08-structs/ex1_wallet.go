package l08structs

// Wallet хранит баланс в копейках.
//
// Задача «найди баг»: методы ниже уже написаны, логика в них верная,
// но Deposit и Withdraw «не работают»: баланс после них не меняется.
// Найди причину и почини, не меняя имён, параметров и логики проверок.
type Wallet struct {
	balance int64
}

// Deposit зачисляет amount копеек. amount <= 0 игнорируется.
// Получатель указатель: метод меняет сам кошелёк, а не его копию.
func (w *Wallet) Deposit(amount int64) {
	if amount <= 0 {
		return
	}
	w.balance += amount
}

// Withdraw списывает amount копеек и возвращает true.
// Если amount <= 0 или денег не хватает, ничего не меняет и возвращает false.
func (w *Wallet) Withdraw(amount int64) bool {
	if amount <= 0 || amount > w.balance {
		return false
	}
	w.balance -= amount
	return true
}

// Balance возвращает текущий баланс.
// Получатель тоже указатель: не смешиваем виды получателей у одного типа.
func (w *Wallet) Balance() int64 {
	return w.balance
}
