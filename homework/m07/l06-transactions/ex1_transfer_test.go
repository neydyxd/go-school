package l06transactions

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/homework/internal/testdb"
)

func balances(t *testing.T, pool *pgxpool.Pool) []string {
	t.Helper()
	var out []string
	for _, r := range testdb.Rows(t, pool, `SELECT balance FROM accounts ORDER BY id`) {
		out = append(out, r[0])
	}
	return out
}

func TestTransfer(t *testing.T) {
	tests := []struct {
		name             string
		from, to, amount int64
		wantErr          error
		want             []string // балансы счетов 1, 2, 3 после вызова
	}{
		{"обычный перевод", 1, 2, 300, nil, []string{"700", "1300", "0"}},
		{"весь баланс до нуля", 2, 3, 1000, nil, []string{"1000", "0", "1000"}},
		{"не хватает денег", 3, 1, 1, ErrInsufficientFunds, []string{"1000", "1000", "0"}},
		{"на копейку больше баланса", 1, 2, 1001, ErrInsufficientFunds, []string{"1000", "1000", "0"}},
		{"нулевая сумма", 1, 2, 0, ErrInvalidAmount, []string{"1000", "1000", "0"}},
		{"отрицательная сумма", 1, 2, -50, ErrInvalidAmount, []string{"1000", "1000", "0"}},
		{"самому себе", 1, 1, 10, ErrSameAccount, []string{"1000", "1000", "0"}},
		{"нет получателя", 1, 99, 10, ErrAccountNotFound, []string{"1000", "1000", "0"}},
		{"нет отправителя", 99, 1, 10, ErrAccountNotFound, []string{"1000", "1000", "0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := newDB(t)
			err := Transfer(testCtx(t), pool, tt.from, tt.to, tt.amount)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Transfer(%d, %d, %d) = %v, want %v", tt.from, tt.to, tt.amount, err, tt.wantErr)
			}
			if got := balances(t, pool); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("балансы после Transfer(%d, %d, %d) = %v, want %v", tt.from, tt.to, tt.amount, got, tt.want)
			}
		})
	}
}

// Соседняя транзакция списала 100 со счёта 1 и ещё не закоммитилась.
// Transfer должен дождаться её и списать свои 50 с уже нового баланса.
func TestTransferNoLostUpdate(t *testing.T) {
	t.Parallel()
	pool := newDB(t)
	h := hold(t, pool, pgx.ReadCommitted, `UPDATE accounts SET balance = balance - 100 WHERE id = 1`)

	c := start(func() error { return Transfer(testCtx(t), pool, 1, 2, 50) })
	waitBlockedOrDone(t, pool, h, c)
	if err := h.commit(); err != nil {
		t.Fatalf("COMMIT соседней транзакции: %v", err)
	}
	if err := c.wait(t, "Transfer"); err != nil {
		t.Fatalf("Transfer(1, 2, 50) = %v, want nil", err)
	}
	want := []string{"850", "1050", "0"}
	if got := balances(t, pool); !reflect.DeepEqual(got, want) {
		t.Errorf("балансы = %v, want %v: потерянное обновление, Transfer записал баланс, вычисленный до чужого COMMIT", got, want)
	}
}

// Счёт 1 заблокирован соседней транзакцией, Transfer ждёт его. Пока он ждёт,
// счёт 2 должен оставаться свободным: иначе встречный перевод даст
// взаимоблокировку. Проверяем оба направления: порядок «сначала from»
// ломается на 2 -> 1, порядок «сначала to» на 1 -> 2.
func TestTransferLockOrder(t *testing.T) {
	for _, dir := range [][2]int64{{2, 1}, {1, 2}} {
		from, to := dir[0], dir[1]
		t.Run(fmt.Sprintf("%d->%d", from, to), func(t *testing.T) {
			t.Parallel()
			pool := newDB(t)
			h := hold(t, pool, pgx.ReadCommitted, `SELECT id FROM accounts WHERE id = 1 FOR UPDATE`)

			c := start(func() error { return Transfer(testCtx(t), pool, from, to, 10) })
			if !waitBlockedOrDone(t, pool, h, c) {
				t.Fatalf("Transfer(%d, %d, 10) завершился (%v), хотя счёт 1 заблокирован: он должен был ждать блокировку", from, to, c.err)
			}
			_, err := h.tx.Exec(testCtx(t), `SELECT id FROM accounts WHERE id = 2 FOR UPDATE NOWAIT`)
			switch {
			case pgCode(err) == "55P03":
				t.Errorf("Transfer(%d, %d) держит блокировку счёта 2, пока ждёт счёт 1: встречный перевод получит deadlock. Блокируй счета по возрастанию id", from, to)
				_ = h.tx.Rollback(testCtx(t)) // ошибка NOWAIT уже прервала транзакцию соседа
			case err != nil:
				t.Fatalf("соседняя транзакция: %v", err)
			default:
				if err := h.commit(); err != nil {
					t.Fatalf("COMMIT соседней транзакции: %v", err)
				}
			}
			if err := c.wait(t, "Transfer"); err != nil {
				t.Fatalf("Transfer(%d, %d, 10) = %v, want nil", from, to, err)
			}
		})
	}
}

// Встречные переводы толпой: ни deadlock, ни потерянных копеек.
func TestTransferConcurrent(t *testing.T) {
	t.Parallel()
	pool := newDB(t)
	const n = 40
	errs := make(chan error, n)
	for i := range n {
		from, to := int64(1), int64(2)
		if i%2 == 1 {
			from, to = to, from
		}
		go func() { errs <- Transfer(testCtx(t), pool, from, to, 10) }()
	}
	for range n {
		if err := <-errs; err != nil {
			t.Errorf("встречные переводы: %v (код %q)", err, pgCode(err))
		}
	}
	want := []string{"1000", "1000", "0"}
	if got := balances(t, pool); !reflect.DeepEqual(got, want) {
		t.Errorf("после 20 переводов 1->2 и 20 переводов 2->1 по 10 балансы = %v, want %v", got, want)
	}
}
