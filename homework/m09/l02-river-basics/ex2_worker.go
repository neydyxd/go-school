package l02riverbasics

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// ErrUserGone: пользователь удалён, пока задача ждала в очереди.
var ErrUserGone = errors.New("пользователь не найден")

// Mailer отправляет письма. В проде SMTP или HTTP API провайдера, в тестах fake.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// WelcomeSubject это тема приветственного письма.
const WelcomeSubject = "Добро пожаловать"

// WelcomeEmailWorker отправляет приветственное письмо.
type WelcomeEmailWorker struct {
	river.WorkerDefaults[WelcomeEmailArgs] // NextRetry, Timeout и Middleware по умолчанию

	Pool   *pgxpool.Pool
	Mailer Mailer
}

// Work выполняет одну задачу. River доставляет задачи «хотя бы один раз», поэтому
// Work обязан быть идемпотентным: повторный запуск не шлёт второе письмо.
//
//  1. SELECT welcome_sent_at FROM users WHERE id = job.Args.UserID.
//     Пользователя нет: верни river.JobCancel(ErrUserGone). Повторять бессмысленно,
//     задача сразу уйдёт в состояние cancelled.
//  2. welcome_sent_at не NULL: письмо уже ушло, верни nil.
//  3. Mailer.Send(ctx, job.Args.Email, WelcomeSubject, любой текст). Ошибку верни
//     как есть (можно обернуть через %w): River повторит задачу позже.
//  4. UPDATE users SET welcome_sent_at = now() WHERE id = ... Строго ПОСЛЕ
//     успешной отправки: если отметить раньше, а письмо не уйдёт, повтор его уже
//     не отправит.
//
// Между шагами 3 и 4 процесс может упасть, и письмо уйдёт дважды. Полностью это
// закрывает только идемпотентный ключ на стороне почтового сервиса (урок 3).
func (w *WelcomeEmailWorker) Work(ctx context.Context, job *river.Job[WelcomeEmailArgs]) error {
	panic("TODO: Work")
}
