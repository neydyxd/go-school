package config

import (
	"strings"
	"testing"
)

func TestLoad(t *testing.T) {
	base := map[string]string{"DATABASE_URL": "postgres://u:p@localhost:5432/db"}
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string // пусто: конфиг валиден
	}{
		{"минимальный конфиг", nil, ""},
		{"нет базы", map[string]string{"DATABASE_URL": ""}, "DATABASE_URL"},
		{"база без хоста", map[string]string{"DATABASE_URL": "postgres://"}, "DATABASE_URL"},
		{"Redis не той схемы", map[string]string{"REDIS_URL": "http://localhost:6379"}, "REDIS_URL"},
		{"нулевой SHUTDOWN_TIMEOUT оборвал бы запросы", map[string]string{"SHUTDOWN_TIMEOUT": "0s"}, "SHUTDOWN_TIMEOUT"},
		{"нулевой DRAIN_DELAY можно", map[string]string{"DRAIN_DELAY": "0s"}, ""},
		{"порт вне диапазона", map[string]string{"PORT": "70000"}, "PORT"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{}
			for k, v := range base {
				env[k] = v
			}
			for k, v := range tt.env {
				env[k] = v
			}
			_, err := Load(func(k string) string { return env[k] })
			switch {
			case tt.wantErr == "" && err != nil:
				t.Errorf("Load(%v) = %v, want nil", tt.env, err)
			case tt.wantErr != "" && (err == nil || !strings.Contains(err.Error(), tt.wantErr)):
				t.Errorf("Load(%v) = %v, want ошибку про %s", tt.env, err, tt.wantErr)
			}
		})
	}
}

func TestLogValueHidesPasswords(t *testing.T) {
	c, err := Load(func(k string) string {
		return map[string]string{
			"DATABASE_URL": "postgres://u:secret1@localhost/db",
			"REDIS_URL":    "redis://:secret2@localhost:6379/0",
		}[k]
	})
	if err != nil {
		t.Fatal(err)
	}
	if s := c.LogValue().String(); strings.Contains(s, "secret1") || strings.Contains(s, "secret2") {
		t.Errorf("LogValue() = %s, пароль попал в лог", s)
	}
}
