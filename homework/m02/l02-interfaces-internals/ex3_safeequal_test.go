package l02interfacesinternals

import "testing"

type withSlice struct{ s []int }

type point struct{ X, Y int }

func TestSafeEqual(t *testing.T) {
	tests := []struct {
		name   string
		a, b   any
		wantEq bool
		wantOK bool
	}{
		{"равные int", 1, 1, true, true},
		{"разные int", 1, 2, false, true},
		{"int и int64", 1, int64(1), false, true},
		{"строки", "go", "go", true, true},
		{"оба nil", nil, nil, true, true},
		{"nil и nil-указатель", nil, (*int)(nil), false, true},
		{"структуры", point{1, 2}, point{1, 2}, true, true},
		{"слайс и int: типы разные, сравнение безопасно", []int{1}, 1, false, true},
		{"два слайса", []int{1}, []int{1}, false, false},
		{"две мапы", map[int]int{}, map[int]int{}, false, false},
		{"две функции", func() {}, func() {}, false, false},
		{"структура со слайсом", withSlice{}, withSlice{}, false, false},
		{"массив интерфейсов со слайсом", [1]any{[]int{1}}, [1]any{[]int{1}}, false, false},
		{"массив интерфейсов с числами", [1]any{1}, [1]any{1}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eq, ok := SafeEqual(tt.a, tt.b)
			if eq != tt.wantEq || ok != tt.wantOK {
				t.Errorf("SafeEqual(%#v, %#v) = %v, %v, want %v, %v", tt.a, tt.b, eq, ok, tt.wantEq, tt.wantOK)
			}
		})
	}
}
