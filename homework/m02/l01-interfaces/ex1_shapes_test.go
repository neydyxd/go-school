package l01interfaces

import (
	"math"
	"testing"
)

// Компилятор проверит, что типы реализуют интерфейс.
var (
	_ Shape = Rect{}
	_ Shape = Circle{}
	_ Shape = (*Rect)(nil)
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestShapes(t *testing.T) {
	tests := []struct {
		name       string
		s          Shape
		area, perm float64
	}{
		{"прямоугольник 3x4", Rect{3, 4}, 12, 14},
		{"квадрат 0x0", Rect{}, 0, 0},
		{"круг r=1", Circle{1}, math.Pi, 2 * math.Pi},
		{"круг r=2", Circle{2}, 4 * math.Pi, 4 * math.Pi},
		{"указатель на прямоугольник", &Rect{2, 5}, 10, 14},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Area(); !near(got, tt.area) {
				t.Errorf("%#v.Area() = %v, want %v", tt.s, got, tt.area)
			}
			if got := tt.s.Perimeter(); !near(got, tt.perm) {
				t.Errorf("%#v.Perimeter() = %v, want %v", tt.s, got, tt.perm)
			}
		})
	}
}

func TestTotalArea(t *testing.T) {
	tests := []struct {
		name string
		in   []Shape
		want float64
	}{
		{"nil", nil, 0},
		{"пусто", []Shape{}, 0},
		{"разные фигуры", []Shape{Rect{3, 4}, Circle{1}, Rect{1, 1}}, 13 + math.Pi},
		{"с nil внутри", []Shape{nil, Rect{2, 2}, nil}, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TotalArea(tt.in); !near(got, tt.want) {
				t.Errorf("TotalArea(%v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestLargest(t *testing.T) {
	tests := []struct {
		name   string
		in     []Shape
		want   Shape
		wantOK bool
	}{
		{"nil", nil, nil, false},
		{"только nil", []Shape{nil, nil}, nil, false},
		{"одна фигура", []Shape{Rect{1, 2}}, Rect{1, 2}, true},
		{"круг больше квадрата, но меньше 1x4", []Shape{Rect{1, 1}, Circle{1}, Rect{1, 4}}, Rect{1, 4}, true},
		{"круг самый большой", []Shape{Rect{1, 3}, Circle{1}}, Circle{1}, true},
		{"при равенстве первая", []Shape{Rect{2, 3}, Rect{3, 2}, Rect{1, 1}}, Rect{2, 3}, true},
		{"nil в начале", []Shape{nil, Rect{1, 1}}, Rect{1, 1}, true},
		{"нулевая площадь тоже фигура", []Shape{Rect{}}, Rect{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Largest(tt.in)
			if got != tt.want || ok != tt.wantOK {
				t.Errorf("Largest(%v) = %v, %v, want %v, %v", tt.in, got, ok, tt.want, tt.wantOK)
			}
		})
	}
}
