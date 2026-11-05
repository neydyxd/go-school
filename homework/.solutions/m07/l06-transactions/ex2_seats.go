package l06transactions

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
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
	// Один оператор = одна транзакция: подзапрос блокирует первое свободное
	// незаблокированное место, UPDATE тут же его занимает.
	var id int64
	err := pool.QueryRow(ctx, `
		UPDATE seats SET user_id = $2
		WHERE id = (
			SELECT id FROM seats
			WHERE event_id = $1 AND user_id IS NULL
			ORDER BY id
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id`, eventID, userID).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, ErrSoldOut
	}
	return id, err
}
