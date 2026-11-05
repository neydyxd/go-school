// Package l07testingservice: домашка урока «Тестирование сервиса».
// Этот файл готов, менять его не нужно: здесь модель задачи и интерфейс репозитория,
// общие для всех заданий.
package l07testingservice

import (
	"context"
	"errors"
	"time"
)

const (
	StatusTodo       = "todo"
	StatusInProgress = "in_progress"
	StatusDone       = "done"
)

// ErrNotFound: задачи нет или она принадлежит другому пользователю.
var ErrNotFound = errors.New("not found")

type Task struct {
	ID        int64     `json:"id"`
	OwnerID   int64     `json:"owner_id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Filter задаёт выборку списка задач.
type Filter struct {
	Status  string // пусто: любой статус
	Limit   int    // <= 0: 20; больше 100: 100
	AfterID int64  // keyset-пагинация: только задачи с id > AfterID
}

// PageSize приводит Limit к допустимому диапазону.
func (f Filter) PageSize() int {
	switch {
	case f.Limit <= 0:
		return 20
	case f.Limit > 100:
		return 100
	}
	return f.Limit
}

// TaskRepo это порт хранилища задач. Все методы видят только задачи владельца ownerID.
type TaskRepo interface {
	// Create сохраняет задачу и возвращает её с присвоенным ID (ID из аргумента
	// игнорируется), Status "" превращается в "todo", CreatedAt заполняется хранилищем.
	Create(ctx context.Context, t Task) (Task, error)
	// Get возвращает задачу владельца или ErrNotFound (нет такой или чужая).
	Get(ctx context.Context, ownerID, id int64) (Task, error)
	// List возвращает задачи владельца, отфильтрованные по f, по возрастанию ID,
	// не больше f.PageSize() штук. Пусто: пустой слайс или nil.
	List(ctx context.Context, ownerID int64, f Filter) ([]Task, error)
	// UpdateStatus меняет статус и возвращает обновлённую задачу или ErrNotFound.
	// Допустимость перехода проверяет сервис, а не репозиторий.
	UpdateStatus(ctx context.Context, ownerID, id int64, status string) (Task, error)
}
