package l03jsonapi

import (
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, 201, map[string]int{"id": 7})
	if w.Code != 201 {
		t.Errorf("WriteJSON: код %d, want 201", w.Code)
	}
	if ct := w.Result().Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("WriteJSON: Content-Type в отправленном ответе %q, want application/json (заголовки ставятся до WriteHeader!)", ct)
	}
	var got map[string]int
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil || got["id"] != 7 {
		t.Errorf("WriteJSON: тело %q, want {\"id\":7}", w.Body.String())
	}
}

func TestWriteError(t *testing.T) {
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantBody   string
	}{
		{"not found", ErrNotFound, 404,
			`{"error":{"code":"not_found","message":"resource not found"}}`},
		{"обёрнутый not found", fmt.Errorf("get user 7: %w", ErrNotFound), 404,
			`{"error":{"code":"not_found","message":"resource not found"}}`},
		{"конфликт", fmt.Errorf("create: %w", ErrConflict), 409,
			`{"error":{"code":"conflict","message":"resource already exists"}}`},
		{"валидация", FieldErrors{"email": "invalid", "name": "required"}, 422,
			`{"error":{"code":"validation_failed","message":"invalid input","fields":{"email":"invalid","name":"required"}}}`},
		{"обёрнутая валидация", fmt.Errorf("signup: %w", FieldErrors{"age": "out_of_range"}), 422,
			`{"error":{"code":"validation_failed","message":"invalid input","fields":{"age":"out_of_range"}}}`},
		{"APIError из DecodeJSON", &APIError{413, "body_too_large", "слишком большое"}, 413,
			`{"error":{"code":"body_too_large","message":"слишком большое"}}`},
		{"обёрнутый APIError", fmt.Errorf("decode: %w", &APIError{400, "invalid_json", "bad"}), 400,
			`{"error":{"code":"invalid_json","message":"bad"}}`},
		{"неизвестная ошибка не утекает", fmt.Errorf("pq: password authentication failed for user \"admin\""), 500,
			`{"error":{"code":"internal","message":"internal server error"}}`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteError(w, tt.err)
			if w.Code != tt.wantStatus {
				t.Errorf("WriteError(%v): код %d, want %d", tt.err, w.Code, tt.wantStatus)
			}
			if ct := w.Result().Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("WriteError(%v): Content-Type %q, want application/json", tt.err, ct)
			}
			var got, want any
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatalf("WriteError(%v): тело не JSON: %q", tt.err, w.Body.String())
			}
			json.Unmarshal([]byte(tt.wantBody), &want)
			if fmt.Sprint(got) != fmt.Sprint(want) {
				t.Errorf("WriteError(%v): тело\n %s\nwant\n %s", tt.err, w.Body.String(), tt.wantBody)
			}
		})
	}
}
