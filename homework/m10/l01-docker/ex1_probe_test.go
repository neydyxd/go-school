package l01docker

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"net/http/httptrace"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// probeWithin вызывает Probe и валит тест, если она не вернулась за limit.
func probeWithin(t *testing.T, ctx context.Context, url string, timeout, limit time.Duration) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- Probe(ctx, url, timeout) }()
	select {
	case err := <-done:
		return err
	case <-time.After(limit):
		t.Fatalf("Probe(%q, %v) не вернулась за %v: таймаут не работает", url, timeout, limit)
		return nil
	}
}

func TestProbeStatus(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		wantErr bool
	}{
		{"200 OK", http.StatusOK, false},
		{"204 No Content тоже успех", http.StatusNoContent, false},
		{"404", http.StatusNotFound, true},
		{"500", http.StatusInternalServerError, true},
		{"503 сервис не готов", http.StatusServiceUnavailable, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/healthz" {
					http.Error(w, "ждали GET /healthz", http.StatusTeapot)
					return
				}
				w.WriteHeader(tt.code)
			}))
			defer srv.Close()

			err := probeWithin(t, context.Background(), srv.URL+"/healthz", time.Second, 5*time.Second)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Probe(статус %d) = nil, want ошибку", tt.code)
				}
				if !strings.Contains(err.Error(), strconv.Itoa(tt.code)) {
					t.Errorf("Probe(статус %d) = %q, want код %d в тексте ошибки", tt.code, err, tt.code)
				}
			} else if err != nil {
				t.Errorf("Probe(статус %d) = %v, want nil", tt.code, err)
			}
		})
	}
}

func TestProbeTimeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	defer srv.Close()
	defer close(release)

	start := time.Now()
	err := probeWithin(t, context.Background(), srv.URL, 200*time.Millisecond, 3*time.Second)
	if err == nil {
		t.Fatal("Probe(зависший сервер) = nil, want ошибку таймаута")
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("Probe вернулась через %v при таймауте 200ms", d)
	}
}

func TestProbeContextCanceled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	defer srv.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := probeWithin(t, ctx, srv.URL, time.Second, 3*time.Second); err == nil {
		t.Error("Probe(отменённый ctx) = nil, want ошибку")
	}
}

func TestProbeConnectionRefused(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	url := "http://" + ln.Addr().String() + "/healthz"
	ln.Close()
	if err := probeWithin(t, context.Background(), url, time.Second, 3*time.Second); err == nil {
		t.Errorf("Probe(%q, порт закрыт) = nil, want ошибку", url)
	}
}

func TestProbeReusesConnection(t *testing.T) {
	var conns atomic.Int32
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(strings.Repeat("ok\n", 100)))
	}))
	srv.Config.ConnState = func(c net.Conn, s http.ConnState) {
		if s == http.StateNew {
			conns.Add(1)
		}
	}
	srv.Start()
	defer srv.Close()

	// Транспорт возвращает соединение в пул асинхронно, уже после Close тела.
	// httptrace сообщает, когда это случилось, поэтому ждём событие, а не спим.
	idle := make(chan error, 1)
	ctx := httptrace.WithClientTrace(context.Background(), &httptrace.ClientTrace{
		PutIdleConn: func(err error) { idle <- err },
	})
	for i := range 5 {
		if err := probeWithin(t, ctx, srv.URL, time.Second, 3*time.Second); err != nil {
			t.Fatalf("Probe = %v, want nil", err)
		}
		select {
		case err := <-idle:
			if err != nil {
				t.Fatalf("проверка %d: соединение не принято в пул: %v", i+1, err)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("проверка %d: соединение не вернулось в пул за 2s: закрой тело ответа и передавай ctx в запрос (http.NewRequestWithContext), иначе трейс не сработает", i+1)
		}
	}
	if n := conns.Load(); n != 1 {
		t.Errorf("5 проверок открыли %d соединений, want 1: закрой тело и используй общий транспорт", n)
	}
}
