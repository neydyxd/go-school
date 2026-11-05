package l05productionserver

import (
	"log/slog"
	"strings"
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadConfigOK(t *testing.T) {
	tests := []struct {
		name string
		env  map[string]string
		want Config
	}{
		{"только обязательное, остальное по умолчанию",
			map[string]string{"DATABASE_URL": "postgres://db"},
			Config{Addr: ":8080", LogLevel: slog.LevelInfo, ShutdownTimeout: 10 * time.Second, DatabaseURL: "postgres://db"}},
		{"всё задано",
			map[string]string{"PORT": "9000", "LOG_LEVEL": "debug", "SHUTDOWN_TIMEOUT": "1m30s", "DATABASE_URL": "x"},
			Config{Addr: ":9000", LogLevel: slog.LevelDebug, ShutdownTimeout: 90 * time.Second, DatabaseURL: "x"}},
		{"уровень в верхнем регистре",
			map[string]string{"LOG_LEVEL": "WARN", "DATABASE_URL": "x"},
			Config{Addr: ":8080", LogLevel: slog.LevelWarn, ShutdownTimeout: 10 * time.Second, DatabaseURL: "x"}},
		{"error и граничные порты",
			map[string]string{"LOG_LEVEL": "Error", "PORT": "65535", "DATABASE_URL": "x"},
			Config{Addr: ":65535", LogLevel: slog.LevelError, ShutdownTimeout: 10 * time.Second, DatabaseURL: "x"}},
		{"порт 1",
			map[string]string{"PORT": "1", "DATABASE_URL": "x"},
			Config{Addr: ":1", LogLevel: slog.LevelInfo, ShutdownTimeout: 10 * time.Second, DatabaseURL: "x"}},
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
		name     string
		env      map[string]string
		wantVars []string // имена, которые должны быть в тексте ошибки
	}{
		{"нет DATABASE_URL", map[string]string{}, []string{"DATABASE_URL"}},
		{"порт не число", map[string]string{"PORT": "http", "DATABASE_URL": "x"}, []string{"PORT"}},
		{"порт 0", map[string]string{"PORT": "0", "DATABASE_URL": "x"}, []string{"PORT"}},
		{"порт 65536", map[string]string{"PORT": "65536", "DATABASE_URL": "x"}, []string{"PORT"}},
		{"неизвестный уровень", map[string]string{"LOG_LEVEL": "verbose", "DATABASE_URL": "x"}, []string{"LOG_LEVEL"}},
		{"таймаут без единиц", map[string]string{"SHUTDOWN_TIMEOUT": "30", "DATABASE_URL": "x"}, []string{"SHUTDOWN_TIMEOUT"}},
		{"таймаут ноль", map[string]string{"SHUTDOWN_TIMEOUT": "0s", "DATABASE_URL": "x"}, []string{"SHUTDOWN_TIMEOUT"}},
		{"таймаут отрицательный", map[string]string{"SHUTDOWN_TIMEOUT": "-5s", "DATABASE_URL": "x"}, []string{"SHUTDOWN_TIMEOUT"}},
		{"все ошибки сразу", map[string]string{"PORT": "-1", "LOG_LEVEL": "loud", "SHUTDOWN_TIMEOUT": "soon"},
			[]string{"PORT", "LOG_LEVEL", "SHUTDOWN_TIMEOUT", "DATABASE_URL"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := LoadConfig(env(tt.env))
			if err == nil {
				t.Fatalf("LoadConfig(%v) error = nil, want ошибку про %v", tt.env, tt.wantVars)
			}
			for _, v := range tt.wantVars {
				if !strings.Contains(err.Error(), v) {
					t.Errorf("LoadConfig(%v) error = %q, want упоминание %s", tt.env, err, v)
				}
			}
		})
	}
}
