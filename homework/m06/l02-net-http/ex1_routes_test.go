package l02nethttp

import (
	"net/http/httptest"
	"testing"
)

func TestRouter(t *testing.T) {
	tests := []struct {
		name, method, path string
		wantCode           int
		wantBody           string // "-" значит не проверять
		wantAllow          string
	}{
		{"корень", "GET", "/", 200, "home", ""},
		{"неизвестный путь не уходит в корень", "GET", "/nope", 404, "-", ""},
		{"пользователь по id", "GET", "/users/42", 200, "user 42", ""},
		{"id в процентной кодировке", "GET", "/users/%D0%B8", 200, "user и", ""},
		{"точный путь важнее wildcard", "GET", "/users/me", 200, "me", ""},
		{"HEAD работает для GET-маршрута", "HEAD", "/users/42", 200, "-", ""},
		{"создание", "POST", "/users", 201, "created", ""},
		{"удаление", "DELETE", "/users/7", 204, "", ""},
		{"чужой метод на /users/{id}", "POST", "/users/42", 405, "-", "DELETE, GET, HEAD"},
		{"чужой метод на /users", "PUT", "/users", 405, "-", "POST"},
		{"GET на /users без id", "GET", "/users", 405, "-", "POST"},
		{"лишний слеш в конце", "GET", "/users/42/", 404, "-", ""},
		{"файл с вложенным путём", "GET", "/files/a/b/c.txt", 200, "file a/b/c.txt", ""},
		{"пустой хвост", "GET", "/files/", 200, "file ", ""},
	}
	mux := NewRouter()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest(tt.method, tt.path, nil))
			if w.Code != tt.wantCode {
				t.Fatalf("%s %s: код %d, want %d (тело %q)", tt.method, tt.path, w.Code, tt.wantCode, w.Body.String())
			}
			if tt.wantBody != "-" && w.Body.String() != tt.wantBody {
				t.Errorf("%s %s: тело %q, want %q", tt.method, tt.path, w.Body.String(), tt.wantBody)
			}
			if got := w.Result().Header.Get("Allow"); got != tt.wantAllow {
				t.Errorf("%s %s: Allow = %q, want %q", tt.method, tt.path, got, tt.wantAllow)
			}
		})
	}
}
