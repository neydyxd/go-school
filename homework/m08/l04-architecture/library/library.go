// Package library это доменный слой: сущности и ошибки предметной области.
// Он ничего не знает ни про HTTP, ни про SQL, поэтому его импортируют все остальные слои.
// Этот файл готов, менять его не нужно.
package library

import "errors"

// Book это книга в библиотеке. Copies: сколько экземпляров можно выдать прямо сейчас.
type Book struct {
	ID     int64
	Title  string
	Author string
	ISBN   string
	Copies int
}

// Доменные ошибки. Слои ниже возвращают их (можно обёрнутыми через %w),
// HTTP-слой переводит их в коды ответа в одном месте.
var (
	ErrNotFound  = errors.New("книга не найдена")
	ErrISBNTaken = errors.New("книга с таким ISBN уже есть")
	ErrNoCopies  = errors.New("свободных экземпляров нет")
)

// ValidationError означает, что входные данные нарушают правила домена.
type ValidationError struct {
	Field   string // имя поля как в API: title, author, isbn, copies
	Message string // по-русски, для человека
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Message }
