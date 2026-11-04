package l03errors

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

func TestHTTPStatus(t *testing.T) {
	deep := fmt.Errorf("handler: %w", fmt.Errorf("service: %w", fmt.Errorf("repo: %w", ErrNotFound)))
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"nil", nil, 200},
		{"голый ErrNotFound", ErrNotFound, 404},
		{"ErrNotFound в три слоя", deep, 404},
		{"ошибка Store.Get", fmt.Errorf("get %q: %w", "k", ErrNotFound), 404},
		{"FieldError обёрнутая", fmt.Errorf("регистрация: %w", &FieldError{"email", "плохой"}), 400},
		{"FieldError внутри Join", errors.Join(errors.New("x"), &FieldError{"age", "мало"}), 400},
		{"NotFound важнее FieldError", errors.Join(&FieldError{"a", "b"}, ErrNotFound), 404},
		{"дедлайн", fmt.Errorf("запрос в БД: %w", context.DeadlineExceeded), 504},
		{"отмена клиентом", fmt.Errorf("запрос: %w", context.Canceled), 499},
		{"текст похож, но это %v", fmt.Errorf("repo: %v", ErrNotFound), 500},
		{"новая ошибка с тем же текстом", errors.New("not found"), 500},
		{"что-то другое", errors.New("disk failure"), 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPStatus(tt.err); got != tt.want {
				t.Errorf("HTTPStatus(%v) = %d, want %d", tt.err, got, tt.want)
			}
		})
	}
}
