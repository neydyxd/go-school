package l06configlogsdocs

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

const (
	testSecret = "s3cr3t-s3cr3t-s3cr3t-s3cr3t-0042" // 32 байта
	testDBURL  = "postgres://app:Pa55word@db.local:5432/tasks?sslmode=disable"
)

func env(kv map[string]string) func(string) string {
	return func(k string) string { return kv[k] }
}

func validEnv() map[string]string {
	return map[string]string{
		"APP_ENV":          "prod",
		"PORT":             "9000",
		"DATABASE_URL":     testDBURL,
		"JWT_SECRET":       testSecret,
		"LOG_LEVEL":        "WARN",
		"SHUTDOWN_TIMEOUT": "15s",
	}
}

func TestLoadValid(t *testing.T) {
	tests := []struct {
		name string
		args []string
		edit func(map[string]string)
		want Config
	}{
		{"всё из окружения", nil, nil,
			Config{"prod", 9000, testDBURL, testSecret, slog.LevelWarn, 15 * time.Second}},
		{"значения по умолчанию", nil, func(m map[string]string) {
			delete(m, "APP_ENV")
			delete(m, "PORT")
			delete(m, "LOG_LEVEL")
			delete(m, "SHUTDOWN_TIMEOUT")
		}, Config{"dev", 8080, testDBURL, testSecret, slog.LevelInfo, 10 * time.Second}},
		{"флаги важнее окружения", []string{"-port", "7000", "-log-level", "debug"}, nil,
			Config{"prod", 7000, testDBURL, testSecret, slog.LevelDebug, 15 * time.Second}},
		{"флаг при пустом окружении", []string{"-port=7001"}, func(m map[string]string) { delete(m, "PORT") },
			Config{"prod", 7001, testDBURL, testSecret, slog.LevelWarn, 15 * time.Second}},
		{"схема postgresql://", nil, func(m map[string]string) { m["DATABASE_URL"] = "postgresql://u@h/db" },
			Config{"prod", 9000, "postgresql://u@h/db", testSecret, slog.LevelWarn, 15 * time.Second}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validEnv()
			if tt.edit != nil {
				tt.edit(m)
			}
			got, err := Load(tt.args, env(m))
			if err != nil {
				t.Fatalf("Load(%v) err = %v, want nil", tt.args, err)
			}
			if got != tt.want {
				t.Errorf("Load(%v) =\n %+v\nwant\n %+v", tt.args, got, tt.want)
			}
		})
	}
}

func TestLoadInvalid(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		edit     func(map[string]string)
		wantVars []string // имена, которые обязаны быть в тексте ошибки
	}{
		{"пустое окружение: обе обязательные сразу", nil, func(m map[string]string) { clear(m) },
			[]string{"DATABASE_URL", "JWT_SECRET"}},
		{"короткий секрет", nil, func(m map[string]string) { m["JWT_SECRET"] = "short-secret" }, []string{"JWT_SECRET"}},
		{"порт не число", nil, func(m map[string]string) { m["PORT"] = "http" }, []string{"PORT"}},
		{"порт ноль", nil, func(m map[string]string) { m["PORT"] = "0" }, []string{"PORT"}},
		{"порт больше 65535", nil, func(m map[string]string) { m["PORT"] = "70000" }, []string{"PORT"}},
		{"плохой порт во флаге", []string{"-port", "-1"}, nil, []string{"PORT"}},
		{"неизвестный уровень", nil, func(m map[string]string) { m["LOG_LEVEL"] = "verbose" }, []string{"LOG_LEVEL"}},
		{"таймаут без единиц", nil, func(m map[string]string) { m["SHUTDOWN_TIMEOUT"] = "10" }, []string{"SHUTDOWN_TIMEOUT"}},
		{"отрицательный таймаут", nil, func(m map[string]string) { m["SHUTDOWN_TIMEOUT"] = "-1s" }, []string{"SHUTDOWN_TIMEOUT"}},
		{"неизвестное окружение", nil, func(m map[string]string) { m["APP_ENV"] = "staging" }, []string{"APP_ENV"}},
		{"не postgres", nil, func(m map[string]string) { m["DATABASE_URL"] = "mysql://root:Pa55word@db/x" }, []string{"DATABASE_URL"}},
		{"адрес без схемы", nil, func(m map[string]string) { m["DATABASE_URL"] = "localhost:5432" }, []string{"DATABASE_URL"}},
		{"три ошибки разом", nil, func(m map[string]string) {
			m["PORT"] = "x"
			m["LOG_LEVEL"] = "loud"
			delete(m, "JWT_SECRET")
		}, []string{"PORT", "LOG_LEVEL", "JWT_SECRET"}},
		{"неизвестный флаг", []string{"-verbose"}, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validEnv()
			if tt.edit != nil {
				tt.edit(m)
			}
			_, err := Load(tt.args, env(m))
			if err == nil {
				t.Fatalf("Load(%v) err = nil, want ошибку", tt.args)
			}
			msg := err.Error()
			for _, v := range tt.wantVars {
				if !strings.Contains(msg, v) {
					t.Errorf("в ошибке %q нет имени переменной %s", msg, v)
				}
			}
			for _, secret := range []string{"Pa55word", "short-secret"} {
				if strings.Contains(msg, secret) {
					t.Errorf("ошибка %q содержит секрет %q", msg, secret)
				}
			}
		})
	}
}
