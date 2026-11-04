package l03functions

import "testing"

func TestCompose(t *testing.T) {
	inc := func(x int) int { return x + 1 }
	double := func(x int) int { return x * 2 }
	square := func(x int) int { return x * x }
	tests := []struct {
		name string
		fs   []func(int) int
		in   int
		want int
	}{
		{"пустой конвейер", nil, 7, 7},
		{"одна функция", []func(int) int{inc}, 7, 8},
		{"порядок слева направо", []func(int) int{inc, double}, 3, 8},
		{"обратный порядок", []func(int) int{double, inc}, 3, 7},
		{"три функции", []func(int) int{inc, double, square}, 1, 16},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Compose(tt.fs...)(tt.in); got != tt.want {
				t.Errorf("Compose(%d функций)(%d) = %d, want %d", len(tt.fs), tt.in, got, tt.want)
			}
		})
	}
}

func TestComposeNoAliasing(t *testing.T) {
	inc := func(x int) int { return x + 1 }
	fs := []func(int) int{inc, inc}
	f := Compose(fs...)
	fs[0] = func(x int) int { return x * 100 } // меняем исходный слайс после сборки
	if got, want := f(1), 3; got != want {
		t.Errorf("после изменения исходного слайса Compose(...)(1) = %d, want %d: конвейер должен скопировать функции", got, want)
	}
}
