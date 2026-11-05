package l02nethttp

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecover(t *testing.T) {
	tests := []struct {
		name     string
		handler  http.HandlerFunc
		wantCode int
		wantBody string
		wantLog  []any
	}{
		{
			"без паники",
			func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "ok") },
			200, "ok", nil,
		},
		{
			"паника со строкой",
			func(w http.ResponseWriter, r *http.Request) { panic("db: password=secret") },
			500, "internal server error\n", []any{"db: password=secret"},
		},
		{
			"паника с ошибкой",
			func(w http.ResponseWriter, r *http.Request) { panic(io.ErrUnexpectedEOF) },
			500, "internal server error\n", []any{io.ErrUnexpectedEOF},
		},
		{
			"nil map: runtime-паника",
			func(w http.ResponseWriter, r *http.Request) {
				var m map[string]int
				m["x"] = 1
			},
			500, "internal server error\n", nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logged []any
			h := Recover(func(v any) { logged = append(logged, v) })(tt.handler)
			w := httptest.NewRecorder()
			func() {
				defer func() {
					if v := recover(); v != nil {
						t.Fatalf("паника вылетела из middleware: %v", v)
					}
				}()
				h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
			}()
			if w.Code != tt.wantCode || w.Body.String() != tt.wantBody {
				t.Errorf("ответ %d %q, want %d %q", w.Code, w.Body.String(), tt.wantCode, tt.wantBody)
			}
			if tt.name == "nil map: runtime-паника" {
				if len(logged) != 1 {
					t.Fatalf("log вызван %d раз, want 1", len(logged))
				}
				if _, ok := logged[0].(error); !ok {
					t.Errorf("log(%v): want runtime.Error", logged[0])
				}
				return
			}
			if fmt.Sprint(logged) != fmt.Sprint(tt.wantLog) || len(logged) != len(tt.wantLog) {
				t.Errorf("log получил %v, want %v", logged, tt.wantLog)
			}
		})
	}
}

func TestRecoverAbortHandler(t *testing.T) {
	var logged []any
	h := Recover(func(v any) { logged = append(logged, v) })(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(http.ErrAbortHandler)
	}))
	var got any
	func() {
		defer func() { got = recover() }()
		h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
	}()
	if err, _ := got.(error); !errors.Is(err, http.ErrAbortHandler) {
		t.Errorf("паника снаружи = %v, want http.ErrAbortHandler проброшен дальше", got)
	}
	if len(logged) != 0 {
		t.Errorf("ErrAbortHandler залогирован: %v, want без лога", logged)
	}
}
