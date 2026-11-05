// Package service: бизнес-логика трекера. Не знает ни про HTTP, ни про SQL:
// зависит только от интерфейсов репозиториев, которые объявлены здесь же,
// на стороне потребителя.
package service

import (
	"context"
	"errors"
	"time"
)

// Ошибки бизнес-уровня. HTTP-слой переводит их в коды ответа (internal/http/errors.go).
var (
	ErrNotFound       = errors.New("not found")       // 404: нет или чужое (IDOR!)
	ErrConflict       = errors.New("conflict")        // 409: email занят, недопустимый переход статуса
	ErrValidation     = errors.New("validation")      // 422: возвращай *ValidationError, errors.Is(err, ErrValidation) для него true
	ErrUnauthorized   = errors.New("unauthorized")    // 401: неверный логин/пароль, плохой refresh
	ErrNotImplemented = errors.New("not implemented") // 501: заглушка, пока ты не написал код
)

// ValidationError это ошибка входных данных с именем поля (как в JSON).
// HTTP-слой отдаёт её как 422 validation_failed с fields: {Field: Message}.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

// Is делает errors.Is(err, ErrValidation) истинным для любой *ValidationError.
func (e *ValidationError) Is(target error) bool { return target == ErrValidation }

func invalid(field, msg string) error { return &ValidationError{Field: field, Message: msg} }

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
)

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
	CreatedAt    time.Time
}

type Project struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"-"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type Task struct {
	ID         int64      `json:"id"`
	ProjectID  int64      `json:"project_id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	AssigneeID *int64     `json:"assignee_id"` // модуль 09: nil, если исполнителя нет
	DueAt      *time.Time `json:"due_at"`      // модуль 09: nil, если срока нет
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// TaskFilter: фильтры и keyset-пагинация списка задач.
type TaskFilter struct {
	Status  string // пусто: любой
	Limit   int    // 1-100, по умолчанию 20
	AfterID int64  // задачи с id > AfterID
}

// Store это репозитории, привязанные к одной транзакции (см. TxRunner).
type Store struct {
	Users    UserRepo
	Refresh  RefreshRepo
	Projects ProjectRepo
	Tasks    TaskRepo
	Jobs     JobQueue // модуль 09: фоновые задачи в той же транзакции
}

// JobQueue ставит фоновые задачи в очередь. Реализация (repository.Jobs) пишет
// их в таблицу River в той же транзакции, что и остальные записи Store:
// задача появится в очереди тогда и только тогда, когда транзакция закоммитится.
// В тестах сервиса подставляй fake, который запоминает вызовы.
type JobQueue interface {
	TaskAssigned(ctx context.Context, taskID, assigneeID int64) error
}

// TxRunner выполняет fn в одной транзакции: все записи через Store либо
// закоммитятся вместе, либо откатятся вместе. Реализация: repository.TxManager,
// в тестах сервиса fake, который просто вызывает fn с fake-репозиториями.
type TxRunner interface {
	InTx(ctx context.Context, fn func(s Store) error) error
}

// TokenPair это ответ на логин и refresh.
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"` // секунды жизни access-токена
}
