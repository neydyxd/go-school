package l08structs

import "testing"

func TestWallet(t *testing.T) {
	tests := []struct {
		name        string
		deposit     int64
		withdraw    int64
		wantOK      bool
		wantBalance int64
	}{
		{"пополнение и списание", 1000, 300, true, 700},
		{"списать всё", 500, 500, true, 0},
		{"не хватает денег", 100, 101, false, 100},
		{"отрицательное пополнение", -50, 10, false, 0},
		{"отрицательное списание", 100, -10, false, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w Wallet
			w.Deposit(tt.deposit)
			ok := w.Withdraw(tt.withdraw)
			if ok != tt.wantOK || w.Balance() != tt.wantBalance {
				t.Errorf("Deposit(%d), Withdraw(%d) = %v, баланс %d; want %v, баланс %d",
					tt.deposit, tt.withdraw, ok, w.Balance(), tt.wantOK, tt.wantBalance)
			}
		})
	}
}

func TestWalletThroughPointer(t *testing.T) {
	w := &Wallet{}
	w.Deposit(250)
	w.Deposit(250)
	if got := w.Balance(); got != 500 {
		t.Errorf("два Deposit(250): Balance() = %d, want 500", got)
	}
}
