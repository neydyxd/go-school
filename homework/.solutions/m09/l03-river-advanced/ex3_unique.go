package l03riveradvanced

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

// DigestArgs: задача «собрать дайджест пользователю». RequestedBy нужен только
// для логов (кто нажал кнопку), в уникальность он входить НЕ должен.
//
// UserID помечен тегом river:"unique": при ByArgs в ключ
// уникальности попадает только он.
type DigestArgs struct {
	UserID      int64  `json:"user_id" river:"unique"`
	RequestedBy string `json:"requested_by"`
}

func (DigestArgs) Kind() string { return "digest" }

// InsertOpts: один дайджест на пользователя в сутки (по UTC):
// UniqueOpts{ByArgs: true, ByPeriod: 24 * time.Hour}.
func (DigestArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{UniqueOpts: river.UniqueOpts{ByArgs: true, ByPeriod: 24 * time.Hour}}
}

// EnqueueDigest ставит дайджест пользователю userID на момент at (InsertOpts
// с ScheduledAt: at; настройки уникальности возьмутся из DigestArgs.InsertOpts).
// Вернуть id задачи и признак «такая уже есть» (JobInsertResult.UniqueSkippedAsDuplicate).
// На дубле River не возвращает ошибку, а отдаёт уже существующую задачу.
func EnqueueDigest(ctx context.Context, client *river.Client[pgx.Tx], userID int64, requestedBy string, at time.Time) (int64, bool, error) {
	res, err := client.Insert(ctx, DigestArgs{UserID: userID, RequestedBy: requestedBy}, &river.InsertOpts{ScheduledAt: at})
	if err != nil {
		return 0, false, err
	}
	return res.Job.ID, res.UniqueSkippedAsDuplicate, nil
}
