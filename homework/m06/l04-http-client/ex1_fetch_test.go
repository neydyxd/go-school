package l04httpclient

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

type item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestFetchJSON(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /ok", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"id":1,"name":"чай"}`)
	})
	mux.HandleFunc("GET /missing", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"not found"}`, http.StatusNotFound)
	})
	mux.HandleFunc("GET /huge-error", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, strings.Repeat("x", 1<<20)) // мегабайт
	})
	mux.HandleFunc("GET /broken", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"id":`)
	})
	srv, client, tr, _ := newTrackedServer(t, mux)

	t.Run("200", func(t *testing.T) {
		var got item
		if err := FetchJSON(context.Background(), client, srv.URL+"/ok", &got); err != nil {
			t.Fatalf("FetchJSON(/ok) = %v, want nil", err)
		}
		if got != (item{1, "чай"}) {
			t.Errorf("FetchJSON(/ok) записал %+v, want {1 чай}", got)
		}
	})
	t.Run("404", func(t *testing.T) {
		var got item
		err := FetchJSON(context.Background(), client, srv.URL+"/missing", &got)
		var se *StatusError
		if !errors.As(err, &se) || se.Code != 404 {
			t.Fatalf("FetchJSON(/missing) = %v, want *StatusError с Code 404", err)
		}
		if !strings.Contains(se.Body, "not found") {
			t.Errorf("StatusError.Body = %q, want начало тела ответа", se.Body)
		}
	})
	t.Run("огромное тело ошибки обрезано", func(t *testing.T) {
		readBefore := tr.read.Load()
		err := FetchJSON(context.Background(), client, srv.URL+"/huge-error", &item{})
		if n := tr.read.Load() - readBefore; n > 64<<10 {
			t.Errorf("из тела ошибки в мегабайт прочитано %d байт, want не больше 64 КБ: читай через io.LimitReader", n)
		}
		var se *StatusError
		if !errors.As(err, &se) || se.Code != 502 {
			t.Fatalf("FetchJSON(/huge-error) = %v, want *StatusError 502", err)
		}
		if len(se.Body) != 512 {
			t.Errorf("len(StatusError.Body) = %d, want 512", len(se.Body))
		}
	})
	t.Run("битый JSON", func(t *testing.T) {
		if err := FetchJSON(context.Background(), client, srv.URL+"/broken", &item{}); err == nil {
			t.Error("FetchJSON(/broken) = nil, want ошибку декодирования")
		}
	})
	tr.checkAllClosed(t)
}

func TestFetchJSONReusesConnection(t *testing.T) {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/err" {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprint(w, "busy")
			return
		}
		fmt.Fprint(w, `{"id":2}`)
	})
	srv, client, tr, conns := newTrackedServer(t, h)
	for i := range 10 {
		path := "/ok"
		if i%2 == 1 {
			path = "/err"
		}
		_ = FetchJSON(context.Background(), client, srv.URL+path, &item{})
	}
	tr.checkAllClosed(t)
	if n := conns.Load(); n != 1 {
		t.Errorf("10 последовательных запросов открыли %d TCP-соединений, want 1: тело не закрыто, соединение не вернулось в пул", n)
	}
}

func TestFetchJSONCancel(t *testing.T) {
	release := make(chan struct{})
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	})
	srv, client, _, _ := newTrackedServer(t, h)
	defer close(release)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- FetchJSON(ctx, client, srv.URL+"/slow", &item{}) }()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("FetchJSON с истёкшим ctx = %v, want context.DeadlineExceeded", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("FetchJSON не вернулся через 2с после отмены ctx: запрос создан без контекста?")
	}
}
