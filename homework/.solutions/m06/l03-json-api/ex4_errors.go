package l03jsonapi

import (
	"encoding/json"
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// ошибку кодирования игнорируем: статус уже отправлен, исправить ответ нельзя
	_ = json.NewEncoder(w).Encode(v)
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Fields  FieldErrors `json:"fields,omitempty"`
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
	var (
		ae     *APIError
		fe     FieldErrors
		status int
		d      errorDetail
	)
	switch {
	case errors.As(err, &ae):
		status, d = ae.Status, errorDetail{Code: ae.Code, Message: ae.Message}
	case errors.As(err, &fe):
		status, d = http.StatusUnprocessableEntity, errorDetail{Code: "validation_failed", Message: "invalid input", Fields: fe}
	case errors.Is(err, ErrNotFound):
		status, d = http.StatusNotFound, errorDetail{Code: "not_found", Message: "resource not found"}
	case errors.Is(err, ErrConflict):
		status, d = http.StatusConflict, errorDetail{Code: "conflict", Message: "resource already exists"}
	default:
		status, d = http.StatusInternalServerError, errorDetail{Code: "internal", Message: "internal server error"}
	}
	WriteJSON(w, status, errorBody{d})
}
