package l01web

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

const appOrigin = "https://app.example.com"

func corsHandler(called *bool) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("data"))
	})
	cors := CORS([]string{appOrigin, "http://localhost:5173"}, next)
	// Внешний middleware (например, gzip) уже поставил свой Vary до CORS.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Accept-Encoding")
		cors.ServeHTTP(w, r)
	})
}

func serve(h http.Handler, method, origin, acrm string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, "/api/items", nil)
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	if acrm != "" {
		r.Header.Set("Access-Control-Request-Method", acrm)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func TestCORSNoOrigin(t *testing.T) {
	var called bool
	w := serve(corsHandler(&called), "GET", "", "")
	if !called || w.Code != 200 {
		t.Fatalf("без Origin: called=%v code=%d, want next вызван и 200", called, w.Code)
	}
	if got := w.Result().Header.Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("без Origin: Access-Control-Allow-Origin = %q, want пусто", got)
	}
	if got := w.Result().Header.Values("Vary"); !slices.Equal(got, []string{"Accept-Encoding"}) {
		t.Errorf("без Origin: Vary = %q, want только то, что поставил внешний middleware", got)
	}
}

func TestCORSSimpleRequest(t *testing.T) {
	tests := []struct {
		name, method, origin string
		wantACAO             string
	}{
		{"разрешённый origin", "GET", appOrigin, appOrigin},
		{"второй разрешённый", "POST", "http://localhost:5173", "http://localhost:5173"},
		{"чужой origin", "GET", "https://evil.example", ""},
		{"origin отличается портом", "GET", "http://localhost:3000", ""},
		{"origin со слешем в конце не тот же", "GET", appOrigin + "/", ""},
		{"OPTIONS без Access-Control-Request-Method это обычный запрос", "OPTIONS", appOrigin, appOrigin},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var called bool
			w := serve(corsHandler(&called), tt.method, tt.origin, "")
			if !called {
				t.Fatalf("%s %s: next не вызван, а CORS не должен блокировать на сервере", tt.method, tt.origin)
			}
			if w.Code != 200 || w.Body.String() != "data" {
				t.Errorf("код %d тело %q, want 200 и ответ next", w.Code, w.Body.String())
			}
			if got := w.Result().Header.Get("Access-Control-Allow-Origin"); got != tt.wantACAO {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantACAO)
			}
			if got := w.Result().Header.Values("Vary"); !slices.Contains(got, "Origin") || !slices.Contains(got, "Accept-Encoding") {
				t.Errorf("Vary = %q, want и Origin, и Accept-Encoding (Add, а не Set)", got)
			}
		})
	}
}

func TestCORSPreflight(t *testing.T) {
	t.Run("разрешённый origin", func(t *testing.T) {
		var called bool
		w := serve(corsHandler(&called), "OPTIONS", appOrigin, "DELETE")
		if called {
			t.Error("preflight дошёл до next, а должен обрабатываться в middleware")
		}
		if w.Code != http.StatusNoContent {
			t.Errorf("код %d, want 204", w.Code)
		}
		want := map[string]string{
			"Access-Control-Allow-Origin":  appOrigin,
			"Access-Control-Allow-Methods": "GET, POST, PUT, PATCH, DELETE",
			"Access-Control-Allow-Headers": "Content-Type, Authorization",
			"Access-Control-Max-Age":       "600",
		}
		for k, v := range want {
			if got := w.Result().Header.Get(k); got != v {
				t.Errorf("%s = %q, want %q", k, got, v)
			}
		}
		if got := w.Result().Header.Values("Vary"); !slices.Contains(got, "Origin") {
			t.Errorf("Vary = %q, want с Origin", got)
		}
	})
	t.Run("чужой origin", func(t *testing.T) {
		var called bool
		w := serve(corsHandler(&called), "OPTIONS", "https://evil.example", "DELETE")
		if called {
			t.Error("preflight дошёл до next")
		}
		if w.Code != http.StatusForbidden {
			t.Errorf("код %d, want 403", w.Code)
		}
		for k := range w.Result().Header {
			if len(k) > 15 && k[:15] == "Access-Control-" {
				t.Errorf("у отказа есть заголовок %s, want ни одного Access-Control-*", k)
			}
		}
		if got := w.Result().Header.Values("Vary"); !slices.Contains(got, "Origin") {
			t.Errorf("Vary = %q, want с Origin", got)
		}
	})
}
