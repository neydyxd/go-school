package l01goquestions

import (
	"errors"
	"maps"
	"testing"
	"time"
)

const base = "https://api.example.com"

func TestNewClientDefaults(t *testing.T) {
	c, err := NewClient(base)
	if err != nil || c == nil {
		t.Fatalf("NewClient(%q) = %v, %v; want клиент, nil", base, c, err)
	}
	if c.BaseURL != base || c.Timeout != 10*time.Second || c.Retries != 2 || c.Headers == nil || len(c.Headers) != 0 {
		t.Errorf("NewClient(%q) = %+v, want BaseURL как передали, Timeout 10s, Retries 2, Headers пустая не nil", base, *c)
	}
}

func TestNewClientOptions(t *testing.T) {
	c, err := NewClient(base,
		WithTimeout(time.Second), WithRetries(0), WithHeader("X-A", "1"),
		WithTimeout(3*time.Second), WithHeader("X-A", "2"), WithHeader("X-B", ""),
	)
	if err != nil {
		t.Fatalf("NewClient с опциями: %v", err)
	}
	want := map[string]string{"X-A": "2", "X-B": ""}
	if c.Timeout != 3*time.Second || c.Retries != 0 || !maps.Equal(c.Headers, want) {
		t.Errorf("NewClient = %+v, want Timeout 3s (поздняя опция), Retries 0, Headers %v", *c, want)
	}
}

func TestNewClientErrors(t *testing.T) {
	tests := []struct {
		name string
		url  string
		opts []Option
		want error
	}{
		{"пустой URL", "", nil, ErrInvalidURL},
		{"без схемы", "api.example.com", nil, ErrInvalidURL},
		{"чужая схема", "ftp://example.com", nil, ErrInvalidURL},
		{"без хоста", "https://", nil, ErrInvalidURL},
		{"мусор", "://bad", nil, ErrInvalidURL},
		{"нулевой таймаут", base, []Option{WithTimeout(0)}, ErrInvalidOption},
		{"отрицательный таймаут", base, []Option{WithTimeout(-time.Second)}, ErrInvalidOption},
		{"отрицательные повторы", base, []Option{WithRetries(-1)}, ErrInvalidOption},
		{"слишком много повторов", base, []Option{WithRetries(11)}, ErrInvalidOption},
		{"пустое имя заголовка", base, []Option{WithHeader("", "x")}, ErrInvalidOption},
		{"ошибка после хорошей опции", base, []Option{WithRetries(3), WithTimeout(0)}, ErrInvalidOption},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := NewClient(tt.url, tt.opts...)
			if c != nil || !errors.Is(err, tt.want) {
				t.Errorf("NewClient(%q) = %v, %v; want nil, ошибку %v", tt.url, c, err, tt.want)
			}
		})
	}
	t.Run("граница повторов 10 допустима", func(t *testing.T) {
		if c, err := NewClient(base, WithRetries(10)); err != nil || c.Retries != 10 {
			t.Errorf("NewClient(WithRetries(10)) = %v, %v; want Retries 10", c, err)
		}
	})
}

func TestNewClientIsolation(t *testing.T) {
	t.Run("заголовки по умолчанию не общие", func(t *testing.T) {
		a, _ := NewClient(base)
		b, _ := NewClient(base)
		a.Headers["X-Leak"] = "1"
		if _, ok := b.Headers["X-Leak"]; ok {
			t.Errorf("заголовок, добавленный клиенту a, появился у b: мапа по умолчанию общая")
		}
	})
	t.Run("одни и те же опции, разные клиенты", func(t *testing.T) {
		opts := []Option{WithHeader("X-A", "1")}
		a, _ := NewClient(base, opts...)
		b, _ := NewClient(base, opts...)
		a.Headers["X-B"] = "2"
		if _, ok := b.Headers["X-B"]; ok {
			t.Errorf("клиенты с одинаковыми опциями делят Headers")
		}
	})
}
