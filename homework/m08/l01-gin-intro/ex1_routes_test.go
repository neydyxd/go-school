package l01ginintro

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() { gin.SetMode(gin.TestMode) }

// serve прогоняет запрос через роутер без сети.
func serve(h http.Handler, method, url string, header map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	h.ServeHTTP(w, req)
	return w
}

// decodeOne требует, чтобы в теле был ровно один JSON-объект.
// Два объекта подряд означают забытый return после ответа с ошибкой.
func decodeOne(t *testing.T, body []byte) map[string]any {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(body))
	var got map[string]any
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("тело не JSON-объект: %q (%v)", body, err)
	}
	if _, err := dec.Token(); err != io.EOF {
		t.Fatalf("в теле больше одного JSON-значения: %q (забыт return после c.JSON?)", body)
	}
	return got
}

type routeCase struct {
	name     string
	url      string
	wantCode int
	want     map[string]any // nil: проверяем только, что есть непустое поле error
}

func checkRoutes(t *testing.T, h http.Handler, tests []routeCase) {
	t.Helper()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := serve(h, http.MethodGet, tt.url, nil)
			if w.Code != tt.wantCode {
				t.Fatalf("GET %s: код %d, want %d (тело %q)", tt.url, w.Code, tt.wantCode, w.Body)
			}
			got := decodeOne(t, w.Body.Bytes())
			if tt.want == nil {
				if s, _ := got["error"].(string); s == "" || len(got) != 1 {
					t.Fatalf("GET %s = %v, want {\"error\": \"<текст>\"}", tt.url, got)
				}
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("GET %s = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestNewRouter(t *testing.T) {
	checkRoutes(t, NewRouter(), []routeCase{
		{"ping", "/ping", 200, map[string]any{"message": "pong"}},
		{"id число", "/users/42", 200, map[string]any{"id": 42.0}},
		{"id не число", "/users/abc", 400, nil},
		{"id ноль", "/users/0", 400, nil},
		{"id отрицательный", "/users/-5", 400, nil},
		{"id переполняет int64", "/users/99999999999999999999", 400, nil},
		{"путь", "/files/docs/a.txt", 200, map[string]any{"path": "docs/a.txt"}},
		{"путь из одного файла", "/files/a", 200, map[string]any{"path": "a"}},
		{"пустой путь", "/files/", 400, nil},
		{"поиск", "/search?q=go&limit=5", 200, map[string]any{"q": "go", "limit": 5.0}},
		{"limit по умолчанию", "/search?q=gin", 200, map[string]any{"q": "gin", "limit": 10.0}},
		{"пробелы в q", "/search?q=%20gin%20", 200, map[string]any{"q": "gin", "limit": 10.0}},
		{"limit на границе", "/search?q=go&limit=50", 200, map[string]any{"q": "go", "limit": 50.0}},
		{"limit на нижней границе", "/search?q=go&limit=1", 200, map[string]any{"q": "go", "limit": 1.0}},
		{"нет q", "/search", 400, nil},
		{"q из пробелов", "/search?q=%20%20", 400, nil},
		{"limit больше 50", "/search?q=go&limit=51", 400, nil},
		{"limit ноль", "/search?q=go&limit=0", 400, nil},
		{"limit не число", "/search?q=go&limit=ten", 400, nil},
	})
}
