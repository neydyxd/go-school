package l06transactions

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrSoldOut = errors.New("свободных мест нет")

// ReserveSeat бронирует для userID одно свободное место (user_id IS NULL)
// на событии eventID и возвращает id места. Берёт свободное место
// с наименьшим id среди тех, что сейчас никто не бронирует.
//
// Сотни покупателей жмут «купить» одновременно, поэтому:
//   - одно место нельзя продать дважды;
//   - покупатель не должен ждать, пока другой покупатель додумает своё
//     место: если строка места заблокирована чужой транзакцией, бери следующее
//     (SELECT ... FOR UPDATE SKIP LOCKED);
//   - если свободных незаблокированных мест нет, верни ErrSoldOut.
func ReserveSeat(ctx context.Context, pool *pgxpool.Pool, eventID, userID int64) (int64, error) {
	panic("TODO: ReserveSeat")
}
