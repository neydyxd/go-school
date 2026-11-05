package l06configlogsdocs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"
)

func TestConfigLogValue(t *testing.T) {
	cfg := Config{"prod", 9000, testDBURL, testSecret, slog.LevelWarn, 15 * time.Second}

	var buf bytes.Buffer
	slog.New(slog.NewJSONHandler(&buf, nil)).Info("старт", "config", cfg)
	out := buf.String()

	var line struct {
		Config map[string]any `json:"config"`
	}
	if err := json.Unmarshal(buf.Bytes(), &line); err != nil {
		t.Fatalf("строка лога не JSON: %v\n%s", err, out)
	}
	want := map[string]any{
		"env":              "prod",
		"port":             float64(9000),
		"database_url":     "postgres://app:xxxxx@db.local:5432/tasks?sslmode=disable",
		"jwt_secret":       "[REDACTED]",
		"log_level":        "WARN",
		"shutdown_timeout": float64(15 * time.Second),
	}
	for k, v := range want {
		if line.Config[k] != v {
			t.Errorf("config.%s = %v, want %v\nстрока: %s", k, line.Config[k], v, out)
		}
	}
	for _, secret := range []string{testSecret, "Pa55word"} {
		if strings.Contains(out, secret) {
			t.Errorf("в лог утёк секрет %q: %s", secret, out)
		}
	}

	t.Run("битый DATABASE_URL", func(t *testing.T) {
		c := cfg
		c.DatabaseURL = "postgres://u:Pa55word@[::1" // url.Parse не разберёт
		s := c.String()
		if strings.Contains(s, "Pa55word") || !strings.Contains(s, "database_url=[REDACTED]") {
			t.Errorf("String() = %s, want database_url=[REDACTED] без пароля", s)
		}
	})

	t.Run("fmt не печатает секреты", func(t *testing.T) {
		for _, verb := range []string{"%v", "%+v", "%s"} {
			s := fmt.Sprintf(verb, cfg)
			if strings.Contains(s, testSecret) || strings.Contains(s, "Pa55word") {
				t.Errorf("fmt.Sprintf(%q, cfg) = %s: секрет в выводе", verb, s)
			}
			if !strings.Contains(s, "port=9000") {
				t.Errorf("fmt.Sprintf(%q, cfg) = %s: want port=9000", verb, s)
			}
		}
	})
}
