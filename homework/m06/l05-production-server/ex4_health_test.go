package l05productionserver

import (
	"context"
	"encoding/json"
	"errors"
	"maps"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"
)

type readyBody struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func callReadyz(t *testing.T, timeout time.Duration, checks map[string]Check) (int, readyBody, *httptest.ResponseRecorder) {
	t.Helper()
	w := httptest.NewRecorder()
	done := make(chan struct{})
	go func() {
		Readyz(timeout, checks).ServeHTTP(w, httptest.NewRequest("GET", "/readyz", nil))
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("/readyz не ответил за 3с при таймауте %v", timeout)
	}
	var b readyBody
	if err := json.Unmarshal(w.Body.Bytes(), &b); err != nil {
		t.Fatalf("тело не JSON: %q", w.Body.String())
	}
	if ct := w.Result().Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if cc := w.Result().Header.Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
	return w.Code, b, w
}

func ok(ctx context.Context) error { return nil }

func TestReadyzResults(t *testing.T) {
	tests := []struct {
		name       string
		checks     map[string]Check
		wantCode   int
		wantStatus string
		wantChecks map[string]string
	}{
		{"без проверок", map[string]Check{}, 200, "ok", map[string]string{}},
		{"все живы", map[string]Check{"db": ok, "redis": ok}, 200, "ok", map[string]string{"db": "ok", "redis": "ok"}},
		{"одна упала", map[string]Check{
			"db":    ok,
			"redis": func(ctx context.Context) error { return errors.New("dial tcp: connection refused") },
		}, 503, "fail", map[string]string{"db": "ok", "redis": "dial tcp: connection refused"}},
		{"проверка честно вернула DeadlineExceeded", map[string]Check{
			"db": func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() },
		}, 503, "fail", map[string]string{"db": "timeout"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, b, w := callReadyz(t, 100*time.Millisecond, tt.checks)
			if code != tt.wantCode || b.Status != tt.wantStatus || !maps.Equal(b.Checks, tt.wantChecks) {
				t.Errorf("ответ %d %s, want %d %+v", code, w.Body.String(), tt.wantCode, readyBody{tt.wantStatus, tt.wantChecks})
			}
		})
	}
}

func TestReadyzParallel(t *testing.T) {
	aStarted, bStarted := make(chan struct{}), make(chan struct{})
	// Каждая проверка ждёт старта другой: последовательно они не пройдут.
	checks := map[string]Check{
		"a": func(ctx context.Context) error {
			close(aStarted)
			select {
			case <-bStarted:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
		"b": func(ctx context.Context) error {
			close(bStarted)
			select {
			case <-aStarted:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		},
	}
	code, b, w := callReadyz(t, time.Second, checks)
	if code != 200 || b.Checks["a"] != "ok" || b.Checks["b"] != "ok" {
		t.Errorf("ответ %d %s, want 200 и обе ok: проверки должны идти параллельно", code, w.Body.String())
	}
}

func TestReadyzHangingCheck(t *testing.T) {
	release := make(chan struct{})
	checks := map[string]Check{
		"db": ok,
		"legacy": func(ctx context.Context) error {
			<-release // не слушает контекст
			return nil
		},
	}
	before := runtime.NumGoroutine()
	start := time.Now()
	code, b, w := callReadyz(t, 100*time.Millisecond, checks)
	if el := time.Since(start); el > time.Second {
		t.Errorf("ответ через %v при таймауте 100мс", el)
	}
	want := map[string]string{"db": "ok", "legacy": "timeout"}
	if code != 503 || b.Status != "fail" || !maps.Equal(b.Checks, want) {
		t.Errorf("ответ %d %s, want 503 и %v", code, w.Body.String(), want)
	}

	// Зависшая проверка наконец вернулась. Её горутина должна отдать
	// результат и завершиться, хотя хендлер давно ушёл.
	close(release)
	deadline := time.Now().Add(time.Second)
	for runtime.NumGoroutine() > before {
		if time.Now().After(deadline) {
			t.Fatalf("горутин %d, до вызова было %d: горутина проверки зависла на отправке результата (небуферизованный канал?)",
				runtime.NumGoroutine(), before)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
