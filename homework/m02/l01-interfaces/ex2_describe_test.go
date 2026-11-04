package l01interfaces

import (
	"errors"
	"testing"
	"time"
)

type both struct{}

func (both) Error() string  { return "я ошибка" }
func (both) String() string { return "я стрингер" }

type color int

func (c color) String() string { return [...]string{"красный", "зелёный"}[c] }

func TestDescribe(t *testing.T) {
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"nil", nil, "nil"},
		{"ошибка", errors.New("диск полон"), "ошибка: диск полон"},
		{"и ошибка, и стрингер", both{}, "ошибка: я ошибка"},
		{"свой стрингер", color(1), "stringer: зелёный"},
		{"time.Duration это стрингер", 90 * time.Second, "stringer: 1m30s"},
		{"int", 42, "int: 42"},
		{"отрицательный int", -7, "int: -7"},
		{"строка латиницей", "go", "строка из 2 рун"},
		{"строка кириллицей", "привет", "строка из 6 рун"},
		{"пустая строка", "", "строка из 0 рун"},
		{"слайс int", []int{1, 2, 3}, "[]int длины 3"},
		{"nil-слайс int", []int(nil), "[]int длины 0"},
		{"bool", true, "bool: true"},
		{"int64 не int", int64(5), "неизвестный тип int64"},
		{"float64", 2.5, "неизвестный тип float64"},
		{"слайс строк", []string{"a"}, "неизвестный тип []string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Describe(tt.in); got != tt.want {
				t.Errorf("Describe(%#v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
