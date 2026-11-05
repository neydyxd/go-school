package l02observability

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"testing/synctest"
	"time"
)

type readyBody struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

// sleepCheck ждёт d, уважая ctx, и возвращает err.
func sleepCheck(d time.Duration, err error) Check {
	return func(ctx context.Context) error {
		if _, ok := ctx.Deadline(); !ok {
			return errors.New("у контекста проверки нет дедлайна")
		}
		select {
		case <-time.After(d):
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// stubbornCheck игнорирует ctx и висит d.
func stubbornCheck(d time.Duration) Check {
	return func(ctx context.Context) error {
		time.Sleep(d)
		return nil
	}
}

func callReady(t *testing.T, h http.Handler) (int, readyBody, time.Duration) {
	t.Helper()
	rec := httptest.NewRecorder()
	start := time.Now()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/readyz", nil))
	elapsed := time.Since(start)
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	var b readyBody
	if err := json.Unmarshal(rec.Body.Bytes(), &b); err != nil {
		t.Fatalf("тело %q не JSON: %v", rec.Body.String(), err)
	}
	return rec.Code, b, elapsed
}

func TestReadiness(t *testing.T) {
	errDown := errors.New("connection refused")
	tests := []struct {
		name     string
		checks   map[string]Check
		wantCode int
		want     readyBody
		wantDur  time.Duration
	}{
		{"без проверок", map[string]Check{}, 200,
			readyBody{"ok", map[string]string{}}, 0},
		{"всё в порядке", map[string]Check{"db": sleepCheck(time.Second, nil), "redis": sleepCheck(time.Second, nil)}, 200,
			readyBody{"ok", map[string]string{"db": "ok", "redis": "ok"}}, time.Second},
		{"одна зависимость упала", map[string]Check{"db": sleepCheck(time.Second, nil), "redis": sleepCheck(time.Second, errDown)}, 503,
			readyBody{"fail", map[string]string{"db": "ok", "redis": "connection refused"}}, time.Second},
		{"проверки идут параллельно", map[string]Check{
			"a": sleepCheck(time.Second, nil), "b": sleepCheck(2*time.Second, nil), "c": sleepCheck(2*time.Second, nil)}, 200,
			readyBody{"ok", map[string]string{"a": "ok", "b": "ok", "c": "ok"}}, 2 * time.Second},
		{"проверка висит и игнорирует ctx", map[string]Check{"db": sleepCheck(time.Second, nil), "legacy": stubbornCheck(time.Hour)}, 503,
			readyBody{"fail", map[string]string{"db": "ok", "legacy": "timeout"}}, 3 * time.Second},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				code, body, dur := callReady(t, Readiness(tt.checks, 3*time.Second))
				if code != tt.wantCode {
					t.Errorf("код = %d, want %d", code, tt.wantCode)
				}
				if !reflect.DeepEqual(body, tt.want) {
					t.Errorf("тело = %+v, want %+v", body, tt.want)
				}
				if dur != tt.wantDur {
					t.Errorf("ответ через %v, want %v", dur, tt.wantDur)
				}
				// Даём зависшим проверкам доработать. Если их горутины не могут
				// отдать результат (небуферизованный канал), synctest упадёт
				// с deadlock: это утечка горутин.
				time.Sleep(2 * time.Hour)
			})
		})
	}
}

func TestReadinessSlowRespectsTimeout(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		h := Readiness(map[string]Check{"db": sleepCheck(time.Minute, nil)}, 2*time.Second)
		code, body, dur := callReady(t, h)
		if code != 503 || body.Status != "fail" {
			t.Errorf("медленная проверка: код %d, status %q, want 503 fail", code, body.Status)
		}
		if v := body.Checks["db"]; v == "ok" || v == "" {
			t.Errorf("checks[db] = %q, want текст ошибки", v)
		}
		if dur != 2*time.Second {
			t.Errorf("ответ через %v, want 2s (таймаут)", dur)
		}
	})
}
