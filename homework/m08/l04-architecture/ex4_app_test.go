package l04architecture

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"goschool/homework/internal/testdb"
)

// Сквозной тест: настоящий роутер, настоящий сервис, настоящая база.
func TestNewRouterEndToEnd(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := NewRouter(testdb.New(t, "testdata/schema.sql"))

	steps := []struct {
		method, url, body string
		status            int
		want              string // подстрока ответа
	}{
		{"POST", "/api/books", `{"title":" Дюна ","author":"Фрэнк Герберт","isbn":"978-0-00-000001-0","copies":1}`, 201,
			`"data":{"id":4,"title":"Дюна","author":"Фрэнк Герберт","isbn":"9780000000010","copies":1}`},
		{"POST", "/api/books", `{"title":"Ещё Дюна","author":"Ф. Г.","isbn":"9780000000010","copies":1}`, 409, `"code":"isbn_taken"`},
		{"POST", "/api/books", `{"title":"","author":"А","isbn":"9780000000011","copies":1}`, 422, `"fields":{"title":"обязательное поле"}`},
		{"POST", "/api/books", `{"title":`, 400, `"code":"bad_request"`},
		{"GET", "/api/books/4", "", 200, `"title":"Дюна"`},
		{"GET", "/api/books/100", "", 404, `"code":"not_found"`},
		{"POST", "/api/books/4/take", "", 200, `"copies":0`},
		{"POST", "/api/books/4/take", "", 409, `"code":"no_copies"`},
		{"POST", "/api/books/100/take", "", 404, `"code":"not_found"`},
	}
	for _, s := range steps {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(s.method, s.url, strings.NewReader(s.body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != s.status || !strings.Contains(w.Body.String(), s.want) || !json.Valid(w.Body.Bytes()) {
			t.Fatalf("%s %s %s: код %d, тело %s; want код %d и %s", s.method, s.url, s.body, w.Code, w.Body, s.status, s.want)
		}
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/books/1", nil))
	if w.Code != http.StatusNotFound {
		t.Fatalf("GET /books/1 без /api: код %d, want 404 (маршруты живут в группе /api)", w.Code)
	}
}
