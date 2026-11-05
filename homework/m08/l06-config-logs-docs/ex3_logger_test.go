package l06configlogsdocs

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// lines разбирает вывод JSON-логгера построчно.
func lines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()
	var out []map[string]any
	for _, l := range strings.Split(strings.TrimSpace(buf.String()), "\n") {
		if l == "" {
			continue
		}
		var m map[string]any
		if err := json.Unmarshal([]byte(l), &m); err != nil {
			t.Fatalf("строка лога не JSON: %q", l)
		}
		out = append(out, m)
	}
	return out
}

func TestNewLoggerLevels(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, slog.LevelInfo)
	l.Debug("не должно попасть")
	l.Info("вход", "user_id", 42)
	l.Error("сбой")
	got := lines(t, &buf)
	if len(got) != 2 {
		t.Fatalf("уровень info: строк %d, want 2 (debug отсекается)\n%s", len(got), buf.String())
	}
	if got[0]["level"] != "INFO" || got[0]["msg"] != "вход" || got[0]["user_id"] != float64(42) {
		t.Errorf("первая строка = %v, want level INFO, msg вход, user_id 42", got[0])
	}
	if _, ok := got[0]["time"]; !ok {
		t.Errorf("в строке нет time: %v", got[0])
	}
	if got[1]["level"] != "ERROR" {
		t.Errorf("вторая строка level = %v, want ERROR", got[1]["level"])
	}

	buf.Reset()
	NewLogger(&buf, slog.LevelDebug).Debug("видно")
	if len(lines(t, &buf)) != 1 {
		t.Errorf("уровень debug: debug-сообщение не записано")
	}
}

func TestNewLoggerRedacts(t *testing.T) {
	var buf bytes.Buffer
	l := NewLogger(&buf, slog.LevelInfo).With("token", "tok-from-with")
	l.Info("вход",
		"user_id", 42,
		"password", "hunter2",
		"Authorization", "Bearer eyJhbGciOi",
		slog.Group("req", "refresh_token", "rt-123", "path", "/v1/login"),
	)
	out := buf.String()
	for _, secret := range []string{"tok-from-with", "hunter2", "eyJhbGciOi", "rt-123"} {
		if strings.Contains(out, secret) {
			t.Errorf("в лог утёк %q:\n%s", secret, out)
		}
	}
	got := lines(t, &buf)[0]
	checks := map[string]any{
		"token":         "[REDACTED]",
		"password":      "[REDACTED]",
		"Authorization": "[REDACTED]",
		"user_id":       float64(42),
	}
	for k, v := range checks {
		if got[k] != v {
			t.Errorf("%s = %v, want %v", k, got[k], v)
		}
	}
	req, _ := got["req"].(map[string]any)
	if req["refresh_token"] != "[REDACTED]" || req["path"] != "/v1/login" {
		t.Errorf("группа req = %v, want refresh_token [REDACTED] и path /v1/login", req)
	}
}
