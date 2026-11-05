package l03jsonapi

import (
	"errors"
	"net/http"
)

// Ошибки доменного слоя. Хендлер о HTTP знает, сервис нет:
// сервис возвращает эти ошибки (часто обёрнутыми через %w), а WriteError
// превращает их в ответ.
var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("conflict")
)

// WriteJSON пишет v как JSON с заголовком Content-Type: application/json
// и статусом status. Порядок важен: заголовки, потом WriteHeader, потом тело.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	panic("TODO: реализуй WriteJSON")
}

// WriteError пишет ошибку в едином формате API:
//
//	{"error":{"code":"not_found","message":"resource not found"}}
//	{"error":{"code":"validation_failed","message":"invalid input","fields":{"email":"invalid"}}}
//
// Поле fields есть только у ошибки валидации. Соответствие (ошибки
// бывают обёрнуты, используй errors.Is и errors.As):
//
//	*APIError                 его Status, Code, Message
//	FieldErrors               422, "validation_failed", "invalid input", fields
//	ErrNotFound               404, "not_found", "resource not found"
//	ErrConflict               409, "conflict", "resource already exists"
//	всё остальное             500, "internal", "internal server error"
//
// Текст неизвестной ошибки клиенту не отдаём: там SQL, пути и прочие детали.
func WriteError(w http.ResponseWriter, err error) {
	panic("TODO: реализуй WriteError")
}
