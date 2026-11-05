package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"goschool/homework/m08/l04-architecture/library"
)

type ctxKey struct{}

// fakeService отвечает заранее заданным результатом и запоминает аргументы.
type fakeService struct {
	book  library.Book
	err   error
	calls []string
	ctxOK bool
}

func (f *fakeService) seen(ctx context.Context, call string) {
	f.calls = append(f.calls, call)
	_, isGin := ctx.(*gin.Context)
	f.ctxOK = !isGin && ctx.Value(ctxKey{}) == "marker"
}

func (f *fakeService) Create(ctx context.Context, title, author, isbn string, copies int) (library.Book, error) {
	f.seen(ctx, fmt.Sprintf("Create(%q, %q, %q, %d)", title, author, isbn, copies))
	return f.book, f.err
}

func (f *fakeService) Get(ctx context.Context, id int64) (library.Book, error) {
	f.seen(ctx, fmt.Sprintf("Get(%d)", id))
	return f.book, f.err
}

func (f *fakeService) Take(ctx context.Context, id int64) (library.Book, error) {
	f.seen(ctx, fmt.Sprintf("Take(%d)", id))
	return f.book, f.err
}

func newRouter(svc BookService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Метка в контексте запроса: сервис должен получить именно c.Request.Context().
	r.Use(func(c *gin.Context) {
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), ctxKey{}, "marker"))
	})
	New(svc).Register(r.Group("/api"))
	return r
}

func do(r http.Handler, method, url, body string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	return w
}

func decodeOne(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	dec := json.NewDecoder(bytes.NewReader(w.Body.Bytes()))
	var got map[string]any
	if err := dec.Decode(&got); err != nil {
		t.Fatalf("тело не JSON-объект: %q", w.Body)
	}
	if _, err := dec.Token(); err != io.EOF {
		t.Fatalf("в теле больше одного JSON-значения: %q (забыт return?)", w.Body)
	}
	return got
}

var dune = library.Book{ID: 7, Title: "Дюна", Author: "Фрэнк Герберт", ISBN: "9780000000010", Copies: 2}

var duneJSON = map[string]any{"data": map[string]any{
	"id": 7.0, "title": "Дюна", "author": "Фрэнк Герберт", "isbn": "9780000000010", "copies": 2.0,
}}

func TestHandlerSuccess(t *testing.T) {
	tests := []struct {
		name, method, url, body string
		status                  int
		call                    string
	}{
		{"создать", "POST", "/api/books", `{"title":"Дюна","author":"Фрэнк Герберт","isbn":"978-0-00-000001-0","copies":2}`, 201,
			`Create("Дюна", "Фрэнк Герберт", "978-0-00-000001-0", 2)`},
		{"получить", "GET", "/api/books/7", "", 200, "Get(7)"},
		{"взять", "POST", "/api/books/7/take", "", 200, "Take(7)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{book: dune}
			w := do(newRouter(svc), tt.method, tt.url, tt.body)
			if w.Code != tt.status {
				t.Fatalf("%s %s: код %d, want %d (тело %q)", tt.method, tt.url, w.Code, tt.status, w.Body)
			}
			if got := decodeOne(t, w); !reflect.DeepEqual(got, duneJSON) {
				t.Fatalf("%s %s = %v, want %v", tt.method, tt.url, got, duneJSON)
			}
			if len(svc.calls) != 1 || svc.calls[0] != tt.call {
				t.Fatalf("вызовы сервиса %v, want [%s]", svc.calls, tt.call)
			}
			if !svc.ctxOK {
				t.Fatal("сервис получил не c.Request.Context(): передавай контекст запроса, а не *gin.Context")
			}
		})
	}
}

func TestHandlerBadRequest(t *testing.T) {
	tests := []struct{ method, url, body string }{
		{"POST", "/api/books", `{"title":`},
		{"POST", "/api/books", `{"title":"Т","copies":"много"}`},
		{"GET", "/api/books/abc", ""},
		{"POST", "/api/books/abc/take", ""},
	}
	for _, tt := range tests {
		t.Run(tt.method+" "+tt.url, func(t *testing.T) {
			svc := &fakeService{book: dune}
			w := do(newRouter(svc), tt.method, tt.url, tt.body)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("код %d, want 400 (тело %q)", w.Code, w.Body)
			}
			e, _ := decodeOne(t, w)["error"].(map[string]any)
			if e == nil || e["code"] != "bad_request" {
				t.Fatalf("тело %q, want {\"error\": {\"code\": \"bad_request\", ...}}", w.Body)
			}
			if len(svc.calls) != 0 {
				t.Fatalf("при плохом запросе вызван сервис: %v", svc.calls)
			}
		})
	}
}

func TestWriteErrorMapping(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		status int
		want   map[string]any
	}{
		{"валидация", &library.ValidationError{Field: "isbn", Message: "ISBN это 13 цифр"}, 422,
			map[string]any{"code": "validation_failed", "message": "проверь поля запроса", "fields": map[string]any{"isbn": "ISBN это 13 цифр"}}},
		{"обёрнутая валидация", fmt.Errorf("create: %w", &library.ValidationError{Field: "title", Message: "обязательное поле"}), 422,
			map[string]any{"code": "validation_failed", "message": "проверь поля запроса", "fields": map[string]any{"title": "обязательное поле"}}},
		{"не найдена", library.ErrNotFound, 404, map[string]any{"code": "not_found", "message": library.ErrNotFound.Error()}},
		{"обёрнутая не найдена", fmt.Errorf("repo: %w", library.ErrNotFound), 404, map[string]any{"code": "not_found", "message": library.ErrNotFound.Error()}},
		{"ISBN занят", library.ErrISBNTaken, 409, map[string]any{"code": "isbn_taken", "message": library.ErrISBNTaken.Error()}},
		{"нет экземпляров", fmt.Errorf("take: %w", library.ErrNoCopies), 409, map[string]any{"code": "no_copies", "message": library.ErrNoCopies.Error()}},
		{"неизвестная", errors.New(`dial tcp: password=hunter2 SELECT * FROM books`), 500,
			map[string]any{"code": "internal", "message": "внутренняя ошибка"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, req := range [][3]string{
				{"POST", "/api/books", `{"title":"Т","author":"А","isbn":"9780000000001","copies":1}`},
				{"GET", "/api/books/1", ""},
				{"POST", "/api/books/1/take", ""},
			} {
				w := do(newRouter(&fakeService{err: tt.err}), req[0], req[1], req[2])
				if w.Code != tt.status {
					t.Fatalf("%s %s: код %d, want %d (тело %q)", req[0], req[1], w.Code, tt.status, w.Body)
				}
				if got := decodeOne(t, w); !reflect.DeepEqual(got, map[string]any{"error": tt.want}) {
					t.Fatalf("%s %s = %v, want {error: %v}", req[0], req[1], got, tt.want)
				}
			}
		})
	}
}
