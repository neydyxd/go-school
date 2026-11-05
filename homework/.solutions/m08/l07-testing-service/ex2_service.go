package l07testingservice

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

var (
	ErrInvalidTitle  = errors.New("title должен быть от 1 до 200 символов")
	ErrInvalidStatus = errors.New("неизвестный статус")
	ErrBadTransition = errors.New("недопустимый переход статуса")
)

// Service это бизнес-логика задач. Зависит только от интерфейса TaskRepo,
// поэтому в тестах вместо базы подставляется fake.
type Service struct {
	repo TaskRepo
}

// NewService готов, менять не нужно.
func NewService(repo TaskRepo) *Service { return &Service{repo: repo} }

// CanTransition сообщает, можно ли перевести задачу из from в to:
//
//	todo        -> in_progress
//	in_progress -> todo, done
//	done        -> in_progress   (переоткрыть)
//
// Всё остальное, включая неизвестные статусы и переход в тот же статус, нельзя.
func CanTransition(from, to string) bool {
	switch from {
	case StatusTodo:
		return to == StatusInProgress
	case StatusInProgress:
		return to == StatusTodo || to == StatusDone
	case StatusDone:
		return to == StatusInProgress
	}
	return false
}

// Create создаёт задачу в статусе todo. title обрезается от пробелов по краям,
// после этого должен быть от 1 до 200 символов (рун, не байт), иначе ErrInvalidTitle
// и в репозиторий ничего не пишется.
func (s *Service) Create(ctx context.Context, ownerID int64, title string) (Task, error) {
	title = strings.TrimSpace(title)
	if n := utf8.RuneCountInString(title); n < 1 || n > 200 {
		return Task{}, ErrInvalidTitle
	}
	return s.repo.Create(ctx, Task{OwnerID: ownerID, Title: title, Status: StatusTodo})
}

// ChangeStatus переводит задачу в статус to.
//
//   - to не один из трёх статусов: ErrInvalidStatus, репозиторий не трогаем вообще;
//   - задачи нет или она чужая: ошибка репозитория, которая через errors.Is совпадает с ErrNotFound;
//   - задача уже в статусе to: вернуть её как есть, UpdateStatus не вызывать (идемпотентность);
//   - переход недопустим (CanTransition): ErrBadTransition, UpdateStatus не вызывать;
//   - иначе UpdateStatus и вернуть его результат.
func (s *Service) ChangeStatus(ctx context.Context, ownerID, id int64, to string) (Task, error) {
	if to != StatusTodo && to != StatusInProgress && to != StatusDone {
		return Task{}, ErrInvalidStatus
	}
	t, err := s.repo.Get(ctx, ownerID, id)
	if err != nil {
		return Task{}, fmt.Errorf("получить задачу %d: %w", id, err)
	}
	if t.Status == to {
		return t, nil
	}
	if !CanTransition(t.Status, to) {
		return Task{}, fmt.Errorf("%w: %s -> %s", ErrBadTransition, t.Status, to)
	}
	return s.repo.UpdateStatus(ctx, ownerID, id, to)
}
