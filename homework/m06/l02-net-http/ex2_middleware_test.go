package l02nethttp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func tag(name string, log *[]string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*log = append(*log, name+">")
			next.ServeHTTP(w, r)
			*log = append(*log, "<"+name)
		})
	}
}

func TestChain(t *testing.T) {
	tests := []struct {
		name string
		mws  []string
		want []string
	}{
		{"без middleware", nil, []string{"h"}},
		{"один", []string{"a"}, []string{"a>", "h", "<a"}},
		{"порядок: первый самый внешний", []string{"a", "b", "c"}, []string{"a>", "b>", "c>", "h", "<c", "<b", "<a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var log []string
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { log = append(log, "h") })
			var mws []Middleware
			for _, n := range tt.mws {
				mws = append(mws, tag(n, &log))
			}
			Chain(h, mws...).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
			if !slices.Equal(log, tt.want) {
				t.Errorf("Chain(%v): порядок вызовов %v, want %v", tt.mws, log, tt.want)
			}
		})
	}
}

func TestRequestID(t *testing.T) {
	tests := []struct {
		name, incoming, want string
	}{
		{"нет входящего: генерируем", "", "gen-1"},
		{"валидный входящий сохраняется", "abc-123-XYZ", "abc-123-XYZ"},
		{"ровно 64 символа", strings.Repeat("a", 64), strings.Repeat("a", 64)},
		{"65 символов: слишком длинный", strings.Repeat("a", 65), "gen-1"},
		{"перевод строки в ID: инъекция в логи", "ok\nlevel=ERROR fake", "gen-1"},
		{"пробел", "a b", "gen-1"},
		{"кириллица", "запрос", "gen-1"},
		{"подчёркивание не разрешено", "a_b", "gen-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seen, headerAtNext string
			h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				seen = RequestIDFrom(r.Context())
				headerAtNext = w.Header().Get("X-Request-Id")
			})
			r := httptest.NewRequest("GET", "/", nil)
			if tt.incoming != "" {
				r.Header["X-Request-Id"] = []string{tt.incoming}
			}
			w := httptest.NewRecorder()
			RequestID(func() string { return "gen-1" })(h).ServeHTTP(w, r)
			if seen != tt.want {
				t.Errorf("входящий %q: RequestIDFrom в хендлере = %q, want %q", tt.incoming, seen, tt.want)
			}
			if headerAtNext != tt.want {
				t.Errorf("входящий %q: X-Request-Id в ответе к моменту next = %q, want %q", tt.incoming, headerAtNext, tt.want)
			}
		})
	}
}

func TestRequestIDFromEmpty(t *testing.T) {
	ctx := context.Background()
	for _, k := range []string{"request_id", "requestID", "requestId", "X-Request-Id", "request-id", "id"} {
		ctx = context.WithValue(ctx, k, "чужое значение")
	}
	if got := RequestIDFrom(ctx); got != "" {
		t.Errorf("RequestIDFrom(контекст без нашего ключа) = %q, want \"\": ключ должен быть своим типом, а не строкой", got)
	}
}
