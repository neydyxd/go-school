package l05context

import (
	"context"
	"testing"
)

func TestRequestID(t *testing.T) {
	bg := context.Background()
	tests := []struct {
		name   string
		ctx    context.Context
		want   string
		wantOK bool
	}{
		{"положили и достали", WithRequestID(bg, "req-42"), "req-42", true},
		{"пустой контекст", bg, "", false},
		{"пустой id не кладём", WithRequestID(bg, ""), "", false},
		{"вложенный перекрывает внешний", WithRequestID(WithRequestID(bg, "a"), "b"), "b", true},
		{"чужие значения поверх не мешают",
			context.WithValue(WithRequestID(bg, "req-1"), otherKey("x"), "y"), "req-1", true},
		{"строковый ключ из чужого пакета не читается как наш",
			context.WithValue(bg, "request_id", "подделка"), "", false},
		{"строковый ключ не перетирает наш",
			context.WithValue(WithRequestID(bg, "настоящий"), "request_id", "подделка"), "настоящий", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := RequestID(tt.ctx)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("RequestID(ctx) = %q, %v, want %q, %v", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

type otherKey string
