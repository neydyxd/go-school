package l03jsonapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
)

// APIError: ошибка, которую можно показать клиенту как есть.
type APIError struct {
	Status  int    // HTTP-статус
	Code    string // машинный код: "invalid_json"
	Message string // человеку
}

func (e *APIError) Error() string { return e.Code + ": " + e.Message }

// DecodeJSON строго читает JSON из тела r в dst (указатель на структуру).
// Ошибки возвращает как *APIError с такими Status и Code:
//
//   - Content-Type не application/json: 415, "unsupported_media_type".
//     Параметры допустимы ("application/json; charset=utf-8"), регистр
//     типа не важен (разбирай через mime.ParseMediaType).
//   - тело длиннее maxBytes: 413, "body_too_large"
//     (http.MaxBytesReader + errors.As с *http.MaxBytesError);
//   - пустое тело: 400, "empty_body";
//   - битый JSON или значение не того типа ("age": "сорок"): 400, "invalid_json";
//   - поле, которого нет в dst: 400, "unknown_field" (опечатка клиента
//     в имени поля не должна молча превращаться в нулевое значение);
//   - после первого JSON-значения есть что-то ещё, кроме пробелов
//     ({"a":1}{"b":2} или {"a":1}xyz): 400, "invalid_json".
//
// Message заполняй на своё усмотрение, тест его не проверяет.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any, maxBytes int64) error {
	mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mt != "application/json" {
		return &APIError{http.StatusUnsupportedMediaType, "unsupported_media_type", "нужен Content-Type: application/json"}
	}

	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var tooBig *http.MaxBytesError
		switch {
		case errors.As(err, &tooBig):
			return &APIError{http.StatusRequestEntityTooLarge, "body_too_large", "тело запроса слишком большое"}
		case errors.Is(err, io.EOF):
			return &APIError{http.StatusBadRequest, "empty_body", "пустое тело запроса"}
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			// у encoding/json нет типизированной ошибки для неизвестного поля, только текст
			return &APIError{http.StatusBadRequest, "unknown_field", err.Error()}
		default:
			return &APIError{http.StatusBadRequest, "invalid_json", err.Error()}
		}
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			return &APIError{http.StatusRequestEntityTooLarge, "body_too_large", "тело запроса слишком большое"}
		}
		return &APIError{http.StatusBadRequest, "invalid_json", "после JSON-объекта лишние данные"}
	}
	return nil
}
