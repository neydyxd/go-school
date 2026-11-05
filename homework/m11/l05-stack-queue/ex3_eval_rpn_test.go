package l05stackqueue

import (
	"errors"
	"testing"
)

func TestEvalRPN(t *testing.T) {
	tests := []struct {
		name    string
		in      []string
		want    int
		wantErr bool
	}{
		{"одно число", []string{"42"}, 42, false},
		{"отрицательное число", []string{"-7"}, -7, false},
		{"сложение и умножение", []string{"2", "1", "+", "3", "*"}, 9, false},
		{"порядок операндов в вычитании", []string{"10", "3", "-"}, 7, false},
		{"порядок операндов в делении", []string{"4", "13", "5", "/", "+"}, 6, false},
		{"деление к нулю", []string{"7", "-2", "/"}, -3, false},
		{"длинное", []string{"10", "6", "9", "3", "+", "-11", "*", "/", "*", "17", "+", "5", "+"}, 22, false},
		{"пустое", nil, 0, true},
		{"не хватает операндов", []string{"1", "+"}, 0, true},
		{"только оператор", []string{"*"}, 0, true},
		{"лишнее число", []string{"1", "2", "3", "+"}, 0, true},
		{"неизвестный токен", []string{"1", "2", "%"}, 0, true},
		{"мусор вместо числа", []string{"1", "x", "+"}, 0, true},
		{"деление на ноль", []string{"1", "0", "/"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvalRPN(tt.in)
			if tt.wantErr {
				if !errors.Is(err, ErrBadExpr) {
					t.Errorf("EvalRPN(%q) = %d, %v, want ошибку ErrBadExpr", tt.in, got, err)
				}
				return
			}
			if err != nil || got != tt.want {
				t.Errorf("EvalRPN(%q) = %d, %v, want %d, nil", tt.in, got, err, tt.want)
			}
		})
	}
}

func TestEvalRPNBig(t *testing.T) {
	// 1 1 + 1 + 1 + ... : стек глубины 2, а вот решение, которое на каждом
	// операторе пересобирает список токенов, будет квадратичным.
	const n = 200_000
	tokens := []string{"1"}
	for range n {
		tokens = append(tokens, "1", "+")
	}
	// и глубокий стек: n единиц, затем n-1 плюсов
	deep := make([]string, 0, 2*n)
	for range n {
		deep = append(deep, "1")
	}
	for range n - 1 {
		deep = append(deep, "+")
	}
	var got1, got2 int
	var err1, err2 error
	within(t, func() {
		got1, err1 = EvalRPN(tokens)
		got2, err2 = EvalRPN(deep)
	})
	if err1 != nil || got1 != n+1 {
		t.Errorf("EvalRPN(цепочка из %d плюсов) = %d, %v, want %d, nil", n, got1, err1, n+1)
	}
	if err2 != nil || got2 != n {
		t.Errorf("EvalRPN(%d единиц и %d плюсов) = %d, %v, want %d, nil", n, n-1, got2, err2, n)
	}
}
