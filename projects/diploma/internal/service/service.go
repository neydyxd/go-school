// Package service содержит бизнес-правила твоей темы. Не знает ни про HTTP, ни про SQL:
// зависит только от интерфейсов репозиториев, которые объявляешь здесь же,
// на стороне потребителя (как в «Трекере задач», m08/l04).
//
// Сюда идут: доменные типы (Booking, Flag, Order), интерфейсы репозиториев,
// сервисы с конструкторами NewXxxService(repo, tx, clock) и юнит-тесты на правила.
package service

import "errors"

// Ошибки бизнес-уровня. HTTP-слой переводит их в коды ответа (internal/http/errors.go),
// репозиторий переводит в них ошибки базы (repository.MapErr).
var (
	ErrNotFound   = errors.New("not found")  // 404: нет или чужое
	ErrConflict   = errors.New("conflict")   // 409: дубль, пересечение брони, недопустимый переход статуса
	ErrValidation = errors.New("validation") // 422: возвращай *ValidationError
)

// ValidationError это ошибка входных данных с именем поля (как в JSON).
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }

// Is делает errors.Is(err, ErrValidation) истинным для любой *ValidationError.
func (e *ValidationError) Is(target error) bool { return target == ErrValidation }
