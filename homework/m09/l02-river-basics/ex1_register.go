package l02riverbasics

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
)

var (
	ErrBadEmail   = errors.New("некорректный email")
	ErrEmailTaken = errors.New("email уже занят")
)

// WelcomeEmailArgs это аргументы задачи «отправить приветственное письмо».
// River сериализует их в JSON и хранит в river_job.args.
type WelcomeEmailArgs struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
}

// Kind это стабильное имя типа задачи, по нему River находит воркер.
// Верни "welcome_email". Менять его после выкатки нельзя: старые задачи в
// таблице останутся со старым именем и не найдут воркер.
func (WelcomeEmailArgs) Kind() string {
	return "" // TODO
}

// InsertOpts задаёт настройки по умолчанию для всех задач этого типа:
// очередь "email" и не больше 5 попыток (MaxAttempts).
func (WelcomeEmailArgs) InsertOpts() river.InsertOpts {
	return river.InsertOpts{} // TODO
}

// RegisterUserTx регистрирует пользователя и ставит приветственное письмо в
// очередь в ПЕРЕДАННОЙ транзакции tx. Commit или Rollback делает вызывающий.
//
//   - email приводится к виду strings.ToLower(strings.TrimSpace(email));
//     пустой или без "@": ErrBadEmail, в базу не ходить;
//   - INSERT INTO users (email) ... RETURNING id; нарушение уникальности
//     (код 23505): ErrEmailTaken;
//   - client.InsertTx(ctx, tx, WelcomeEmailArgs{UserID: id, Email: email}, nil).
//
// Именно InsertTx, а не Insert: задача должна появиться в очереди тогда и
// только тогда, когда закоммитится пользователь. Вернуть id пользователя.
func RegisterUserTx(ctx context.Context, tx pgx.Tx, client *river.Client[pgx.Tx], email string) (int64, error) {
	panic("TODO: RegisterUserTx")
}
