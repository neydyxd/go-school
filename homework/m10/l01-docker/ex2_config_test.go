package l01docker

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func env(kv map[string]string) func(string) string {
	return func(k string) string { return kv[k] }
}

const dsn = "postgres://app:secret@db:5432/app?sslmode=disable"

func TestLoadConfigOK(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Config
	}{
		{"только обязательное, остальное по умолчанию",
			map[string]string{"DATABASE_URL": dsn},
			Config{Port: 8080, DatabaseURL: dsn, LogLevel: slog.LevelInfo, ShutdownTimeout: 10 * time.Second}},
		{"всё задано",
			map[string]string{"DATABASE_URL": dsn, "PORT": "9090", "LOG_LEVEL": "debug", "SHUTDOWN_TIMEOUT": "25s"},
			Config{Port: 9090, DatabaseURL: dsn, LogLevel: slog.LevelDebug, ShutdownTimeout: 25 * time.Second}},
		{"уровень в верхнем регистре",
			map[string]string{"DATABASE_URL": dsn, "LOG_LEVEL": "WARN"},
			Config{Port: 8080, DatabaseURL: dsn, LogLevel: slog.LevelWarn, ShutdownTimeout: 10 * time.Second}},
		{"error и составная длительность",
			map[string]string{"DATABASE_URL": dsn, "LOG_LEVEL": "Error", "SHUTDOWN_TIMEOUT": "1m30s"},
			Config{Port: 8080, DatabaseURL: dsn, LogLevel: slog.LevelError, ShutdownTimeout: 90 * time.Second}},
		{"границы порта: 1",
			map[string]string{"DATABASE_URL": dsn, "PORT": "1"},
			Config{Port: 1, DatabaseURL: dsn, LogLevel: slog.LevelInfo, ShutdownTimeout: 10 * time.Second}},
		{"границы порта: 65535",
			map[string]string{"DATABASE_URL": dsn, "PORT": "65535"},
			Config{Port: 65535, DatabaseURL: dsn, LogLevel: slog.LevelInfo, ShutdownTimeout: 10 * time.Second}},
		{"пустая строка значит не задано",
			map[string]string{"DATABASE_URL": dsn, "PORT": "", "LOG_LEVEL": "", "SHUTDOWN_TIMEOUT": ""},
			Config{Port: 8080, DatabaseURL: dsn, LogLevel: slog.LevelInfo, ShutdownTimeout: 10 * time.Second}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadConfig(env(tt.env))
			if err != nil {
				t.Fatalf("LoadConfig(%v) error = %v, want nil", tt.env, err)
			}
			if got != tt.want {
				t.Errorf("LoadConfig(%v) = %+v, want %+v", tt.env, got, tt.want)
			}
		})
	}
}

func TestLoadConfigErrors(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want []string // имена переменных, которые обязаны быть в тексте ошибки
	}{
		{"нет DATABASE_URL", map[string]string{}, []string{"DATABASE_URL"}},
		{"порт не число", map[string]string{"DATABASE_URL": dsn, "PORT": "http"}, []string{"PORT"}},
		{"порт 0", map[string]string{"DATABASE_URL": dsn, "PORT": "0"}, []string{"PORT"}},
		{"порт 65536", map[string]string{"DATABASE_URL": dsn, "PORT": "65536"}, []string{"PORT"}},
		{"порт отрицательный", map[string]string{"DATABASE_URL": dsn, "PORT": "-80"}, []string{"PORT"}},
		{"неизвестный уровень", map[string]string{"DATABASE_URL": dsn, "LOG_LEVEL": "verbose"}, []string{"LOG_LEVEL"}},
		{"уровень со смещением", map[string]string{"DATABASE_URL": dsn, "LOG_LEVEL": "warn+1"}, []string{"LOG_LEVEL"}},
		{"числовой уровень", map[string]string{"DATABASE_URL": dsn, "LOG_LEVEL": "INFO-4"}, []string{"LOG_LEVEL"}},
		{"таймаут без единиц", map[string]string{"DATABASE_URL": dsn, "SHUTDOWN_TIMEOUT": "10"}, []string{"SHUTDOWN_TIMEOUT"}},
		{"таймаут ноль", map[string]string{"DATABASE_URL": dsn, "SHUTDOWN_TIMEOUT": "0s"}, []string{"SHUTDOWN_TIMEOUT"}},
		{"таймаут отрицательный", map[string]string{"DATABASE_URL": dsn, "SHUTDOWN_TIMEOUT": "-5s"}, []string{"SHUTDOWN_TIMEOUT"}},
		{"все ошибки разом",
			map[string]string{"PORT": "eighty", "LOG_LEVEL": "loud", "SHUTDOWN_TIMEOUT": "soon"},
			[]string{"DATABASE_URL", "PORT", "LOG_LEVEL", "SHUTDOWN_TIMEOUT"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadConfig(env(tt.env))
			if err == nil {
				t.Fatalf("LoadConfig(%v) error = nil, want ошибку про %v", tt.env, tt.want)
			}
			for _, name := range tt.want {
				if !strings.Contains(err.Error(), name) {
					t.Errorf("LoadConfig(%v) error = %q, want упоминание %s", tt.env, err, name)
				}
			}
		})
	}
}
