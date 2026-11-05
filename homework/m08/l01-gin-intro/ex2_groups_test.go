package l01ginintro

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNewAPI(t *testing.T) {
	checkRoutes(t, NewAPI(), []routeCase{
		{"v1", "/api/v1/users/42", 200, map[string]any{"id": "42", "version": 1.0}},
		{"v2", "/api/v2/users/7", 200, map[string]any{"data": map[string]any{"id": "7"}, "version": 2.0}},
		{"нет такой версии", "/api/v3/users/7", 404, map[string]any{"error": "not found"}},
		{"неизвестный путь", "/nope", 404, map[string]any{"error": "not found"}},
	})
}

func TestNewAPIMethodNotAllowed(t *testing.T) {
	w := serve(NewAPI(), http.MethodPost, "/api/v1/users/1", nil)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /api/v1/users/1: код %d, want 405 (включи HandleMethodNotAllowed)", w.Code)
	}
	got := decodeOne(t, w.Body.Bytes())
	if want := map[string]any{"error": "method not allowed"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("POST /api/v1/users/1 = %v, want %v", got, want)
	}
	if allow := w.Header().Get("Allow"); allow != "GET" {
		t.Errorf("заголовок Allow = %q, want \"GET\"", allow)
	}
}

func TestNewAPIAdmin(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		wantCode int
		want     map[string]any
	}{
		{"верный токен", AdminToken, 200, map[string]any{"users": 2.0}},
		{"без токена", "", 403, map[string]any{"error": "forbidden"}},
		{"чужой токен", "guess", 403, map[string]any{"error": "forbidden"}},
	}
	api := NewAPI()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := map[string]string{}
			if tt.token != "" {
				h["X-Admin-Token"] = tt.token
			}
			w := serve(api, http.MethodGet, "/admin/stats", h)
			if w.Code != tt.wantCode {
				t.Fatalf("GET /admin/stats: код %d, want %d (тело %q)", w.Code, tt.wantCode, w.Body)
			}
			// decodeOne упадёт, если после 403 выполнился обработчик и дописал статистику.
			if got := decodeOne(t, w.Body.Bytes()); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GET /admin/stats = %v, want %v", got, tt.want)
			}
		})
	}
}
