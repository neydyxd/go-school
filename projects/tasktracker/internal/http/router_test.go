package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"goschool/tasktracker/internal/service"
)

// Образец HTTP-теста: настоящий роутер, fake вместо сервисов, httptest вместо сети.

func init() { gin.SetMode(gin.TestMode) }

type fakePinger struct{ err error }

func (f fakePinger) Ping(context.Context) error { return f.err }

// fakeAuth реализует AuthService; нужен только Register, остальное не вызывается.
type fakeAuth struct {
	AuthService
	err error
}

func (f fakeAuth) Register(_ context.Context, email, _ string) (service.User, error) {
	if f.err != nil {
		return service.User{}, f.err
	}
	return service.User{ID: 7, Email: email, PasswordHash: "secret-hash"}, nil
}

func newTestRouter(d Deps) *gin.Engine {
	d.Log = slog.New(slog.NewJSONHandler(io.Discard, nil))
	if d.DB == nil {
		d.DB = fakePinger{}
	}
	return NewRouter(d)
}

func do(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealth(t *testing.T) {
	for _, tt := range []struct {
		name string
		err  error
		want int
	}{{"база жива", nil, 200}, {"база лежит", errors.New("dial tcp: refused"), 503}} {
		t.Run(tt.name, func(t *testing.T) {
			w := do(newTestRouter(Deps{DB: fakePinger{tt.err}}), "GET", "/healthz", "")
			if w.Code != tt.want {
				t.Errorf("GET /healthz = %d, want %d", w.Code, tt.want)
			}
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		svcErr     error
		wantCode   int
		wantErr    string // код ошибки в едином формате
		wantFields map[string]string
	}{
		{"успех", `{"email":"a@example.com","password":"password1"}`, nil, 201, "", nil},
		{"битый JSON", `{"email":`, nil, 400, "bad_request", nil},
		{"нет пароля", `{"email":"a@example.com"}`, nil, 422, "validation_failed", map[string]string{"password": "обязательное поле"}},
		{"сервис: валидация", `{"email":"x","password":"p"}`, &service.ValidationError{Field: "email", Message: "некорректный email"},
			422, "validation_failed", map[string]string{"email": "некорректный email"}},
		{"сервис: email занят", `{"email":"a@example.com","password":"password1"}`, service.ErrConflict, 409, "conflict", nil},
		{"сервис: база упала", `{"email":"a@example.com","password":"password1"}`, errors.New("conn reset"), 500, "internal", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := do(newTestRouter(Deps{Auth: fakeAuth{err: tt.svcErr}}), "POST", "/v1/auth/register", tt.body)
			if w.Code != tt.wantCode {
				t.Fatalf("код %d, want %d; тело %s", w.Code, tt.wantCode, w.Body)
			}
			if strings.Contains(w.Body.String(), "secret-hash") || strings.Contains(w.Body.String(), "conn reset") {
				t.Errorf("в ответ утекли внутренности: %s", w.Body)
			}
			var body struct {
				ID    int64
				Error *ErrorBody
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatalf("тело не JSON: %s", w.Body)
			}
			if tt.wantErr == "" {
				if body.ID != 7 {
					t.Errorf("id = %d, want 7", body.ID)
				}
				return
			}
			if body.Error == nil || body.Error.Code != tt.wantErr {
				t.Fatalf("ошибка = %+v, want code %q", body.Error, tt.wantErr)
			}
			if fmt.Sprint(body.Error.Fields) != fmt.Sprint(tt.wantFields) {
				t.Errorf("fields = %v, want %v", body.Error.Fields, tt.wantFields)
			}
		})
	}
}

func TestProtectedRoutesNeedAuth(t *testing.T) {
	// Пока authenticate не реализован, здесь 501. Когда реализуешь, поменяй ожидание на 401
	// и добавь кейсы с валидным токеном (fake TokenParser).
	w := do(newTestRouter(Deps{}), "GET", "/v1/projects", "")
	if w.Code != http.StatusNotImplemented && w.Code != http.StatusUnauthorized {
		t.Errorf("GET /v1/projects без токена = %d, want 401 (или 501, пока не реализовано)", w.Code)
	}
}
