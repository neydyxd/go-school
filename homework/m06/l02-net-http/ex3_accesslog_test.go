package l02nethttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccessLog(t *testing.T) {
	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    Entry
	}{
		{
			"только Write: статус 200",
			func(w http.ResponseWriter, r *http.Request) { fmt.Fprint(w, "hello") },
			Entry{Method: "GET", Path: "/x", Status: 200, Bytes: 5},
		},
		{
			"ничего не записал: всё равно 200",
			func(w http.ResponseWriter, r *http.Request) {},
			Entry{Method: "GET", Path: "/x", Status: 200, Bytes: 0},
		},
		{
			"явный статус",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				fmt.Fprint(w, "ok")
			},
			Entry{Method: "GET", Path: "/x", Status: 201, Bytes: 2},
		},
		{
			"второй WriteHeader не считается",
			func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.WriteHeader(http.StatusInternalServerError)
			},
			Entry{Method: "GET", Path: "/x", Status: 404},
		},
		{
			"WriteHeader после Write не меняет статус",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "partial")
				w.WriteHeader(http.StatusInternalServerError)
			},
			Entry{Method: "GET", Path: "/x", Status: 200, Bytes: 7},
		},
		{
			"несколько Write складываются, кириллица в байтах",
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprint(w, "при")
				fmt.Fprint(w, "вет")
			},
			Entry{Method: "GET", Path: "/x", Status: 200, Bytes: 12},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []Entry
			h := AccessLog(func(e Entry) { got = append(got, e) })(tt.handler)
			h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/x", nil))
			if len(got) != 1 {
				t.Fatalf("log вызван %d раз, want 1", len(got))
			}
			if got[0] != tt.want {
				t.Errorf("Entry = %+v, want %+v", got[0], tt.want)
			}
		})
	}
}

func TestAccessLogClientSeesStatus(t *testing.T) {
	h := AccessLog(func(Entry) {})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test", "1")
		w.WriteHeader(http.StatusTeapot)
		fmt.Fprint(w, "tea")
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if w.Code != http.StatusTeapot || w.Body.String() != "tea" || w.Result().Header.Get("X-Test") != "1" {
		t.Errorf("клиент получил %d %q X-Test=%q, want 418 \"tea\" X-Test=1: обёртка должна пробрасывать всё дальше",
			w.Code, w.Body.String(), w.Result().Header.Get("X-Test"))
	}
}

func TestAccessLogPattern(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /users/{id}", func(w http.ResponseWriter, r *http.Request) {})
	var got Entry
	h := AccessLog(func(e Entry) { got = e })(mux)

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/users/42", nil))
	want := Entry{Method: "GET", Path: "/users/42", Pattern: "GET /users/{id}", Status: 200}
	if got != want {
		t.Errorf("Entry = %+v, want %+v", got, want)
	}

	h.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/nope", nil))
	if got.Pattern != "" || got.Status != 404 {
		t.Errorf("для 404: Pattern=%q Status=%d, want \"\" и 404", got.Pattern, got.Status)
	}
}

func TestAccessLogFlush(t *testing.T) {
	var flushErr error
	h := AccessLog(func(Entry) {})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "chunk")
		flushErr = http.NewResponseController(w).Flush()
	}))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
	if flushErr != nil {
		t.Errorf("Flush через обёртку: %v; want nil (нужен метод Unwrap)", flushErr)
	}
	if !w.Flushed {
		t.Error("Flush не дошёл до исходного ResponseWriter")
	}
}
