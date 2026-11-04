package l03testabledesign

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestFetchProfile(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Accept") != "application/json" {
			http.Error(w, "нужен заголовок Accept: application/json", http.StatusNotAcceptable)
			return
		}
		switch r.URL.Path {
		case "/profiles/42":
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"id": 42, "name": "Гофер"}`)
		case "/profiles/7":
			fmt.Fprint(w, `{"id": 7, "name": `) // обрезанный JSON
		case "/profiles/500":
			http.Error(w, "boom", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	t.Run("профиль найден", func(t *testing.T) {
		c, spy := spyClient(t, srv)
		got, err := FetchProfile(c, srv.URL, 42)
		want := Profile{ID: 42, Name: "Гофер"}
		if err != nil || got != want {
			t.Errorf("FetchProfile(42) = %+v, %v, want %+v, nil", got, err, want)
		}
		spy.check(t, 1)
	})
	t.Run("404", func(t *testing.T) {
		c, spy := spyClient(t, srv)
		_, err := FetchProfile(c, srv.URL, 1)
		if !errors.Is(err, ErrNoProfile) {
			t.Errorf("FetchProfile(1): ошибка %v, want ErrNoProfile", err)
		}
		spy.check(t, 1)
	})
	t.Run("500", func(t *testing.T) {
		c, spy := spyClient(t, srv)
		_, err := FetchProfile(c, srv.URL, 500)
		if err == nil || !strings.Contains(err.Error(), "500") || errors.Is(err, ErrNoProfile) {
			t.Errorf("FetchProfile(500): ошибка %v, want ошибку с кодом 500", err)
		}
		spy.check(t, 1)
	})
	t.Run("битый JSON", func(t *testing.T) {
		c, spy := spyClient(t, srv)
		if _, err := FetchProfile(c, srv.URL, 7); err == nil {
			t.Error("FetchProfile(7) с обрезанным JSON: ошибки нет, want ошибку")
		}
		spy.check(t, 1)
	})
	t.Run("сервер недоступен", func(t *testing.T) {
		dead := httptest.NewServer(http.NotFoundHandler())
		url := dead.URL
		dead.Close()
		c, spy := spyClient(t, dead)
		if _, err := FetchProfile(c, url, 42); err == nil {
			t.Error("FetchProfile на закрытый сервер: ошибки нет, want ошибку")
		}
		if n := spy.requests.Load(); n != 1 {
			t.Errorf("запросов через переданный клиент: %d, want 1", n)
		}
	})
}

// spyTransport пропускает запросы к настоящему транспорту тестового сервера,
// считает их и следит, закрыто ли тело каждого ответа.
type spyTransport struct {
	next     http.RoundTripper
	requests atomic.Int32
	bodies   atomic.Int32
	closed   atomic.Int32
}

func (s *spyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	s.requests.Add(1)
	resp, err := s.next.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	s.bodies.Add(1)
	resp.Body = &closeSpy{ReadCloser: resp.Body, closed: &s.closed}
	return resp, nil
}

type closeSpy struct {
	io.ReadCloser
	closed *atomic.Int32
	once   sync.Once
}

func (c *closeSpy) Close() error {
	c.once.Do(func() { c.closed.Add(1) })
	return c.ReadCloser.Close()
}

// spyClient возвращает клиента со шпионским транспортом поверх srv.Client().
func spyClient(t *testing.T, srv *httptest.Server) (*http.Client, *spyTransport) {
	t.Helper()
	spy := &spyTransport{next: srv.Client().Transport}
	return &http.Client{Transport: spy}, spy
}

// check проверяет, что запрос прошёл через переданный клиент и тело закрыто.
func (s *spyTransport) check(t *testing.T, wantRequests int32) {
	t.Helper()
	if n := s.requests.Load(); n != wantRequests {
		t.Errorf("запросов через переданный клиент c: %d, want %d (используй c.Do, а не http.Get или http.DefaultClient)", n, wantRequests)
	}
	if s.closed.Load() != s.bodies.Load() {
		t.Errorf("тело ответа не закрыто: закрыто %d из %d, нужен defer resp.Body.Close()", s.closed.Load(), s.bodies.Load())
	}
}
