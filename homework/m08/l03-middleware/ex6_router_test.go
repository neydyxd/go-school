package l03middleware

import (
	"net/http"
	"reflect"
	"testing"
)

var tokens = map[string]string{"t-alice": "alice"}

func TestNewRouterRoutes(t *testing.T) {
	log, read := capture(t)
	r := NewRouter(log, tokens, []string{allowed})

	tests := []struct {
		name   string
		url    string
		header map[string]string
		status int
		want   map[string]any
	}{
		{"health без токена", "/health", nil, 200, map[string]any{"status": "ok"}},
		{"me с токеном", "/api/me", map[string]string{"Authorization": "Bearer t-alice"}, 200, map[string]any{"user": "alice"}},
		{"me без токена", "/api/me", nil, 401, map[string]any{"error": "unauthorized"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serve(r, http.MethodGet, tt.url, tt.header)
			if w.Code != tt.status {
				t.Fatalf("GET %s: код %d, want %d", tt.url, w.Code, tt.status)
			}
			if got := decodeOne(t, w.Body.Bytes()); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GET %s = %v, want %v", tt.url, got, tt.want)
			}
			rec := requestRecord(t, read())
			if rec["route"] != tt.url || rec["status"] != float64(tt.status) {
				t.Fatalf("GET %s: в логе route=%v status=%v", tt.url, rec["route"], rec["status"])
			}
			if id := w.Header().Get(HeaderRequestID); id == "" || rec["request_id"] != id {
				t.Fatalf("GET %s: request_id в логе %v, в ответе %q, want одинаковые и непустые",
					tt.url, rec["request_id"], id)
			}
		})
	}
}

func TestNewRouterPanic(t *testing.T) {
	log, read := capture(t)
	r := NewRouter(log, tokens, []string{allowed})
	w := serve(r, http.MethodGet, "/api/panic", map[string]string{"Authorization": "Bearer t-alice"})
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("GET /api/panic: код %d, want 500", w.Code)
	}
	if got, want := decodeOne(t, w.Body.Bytes()), map[string]any{"error": "internal"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("GET /api/panic = %v, want %v", got, want)
	}
	recs := read()
	rec := requestRecord(t, recs) // Recovery снаружи Logger: паника пролетит мимо записи "request"
	if rec["status"] != 500.0 || rec["level"] != "ERROR" {
		t.Fatalf("запись request после паники: %v, want status 500 и уровень ERROR", rec)
	}
	if id := w.Header().Get(HeaderRequestID); id == "" || rec["request_id"] != id {
		t.Fatalf("request_id в логе %v, в ответе %q, want одинаковые", rec["request_id"], id)
	}
}

func TestNewRouterPreflight(t *testing.T) {
	log, _ := capture(t)
	r := NewRouter(log, tokens, []string{allowed})
	w := serve(r, http.MethodOptions, "/api/me", map[string]string{"Origin": allowed, "Access-Control-Request-Method": "GET"})
	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight OPTIONS /api/me без токена: код %d, want 204 (CORS глобально и раньше авторизации)", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != allowed {
		t.Fatalf("preflight: Access-Control-Allow-Origin = %q, want %q", w.Header().Get("Access-Control-Allow-Origin"), allowed)
	}
}
