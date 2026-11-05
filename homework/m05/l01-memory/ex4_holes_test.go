package l01memory

import (
	"reflect"
	"slices"
	"testing"
	"unsafe"
)

func TestHoles(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("ожидаемые значения посчитаны для 64-битной платформы")
	}
	type inner struct {
		X int32
		Y bool
	}
	tests := []struct {
		name string
		typ  reflect.Type
		want []Hole
	}{
		{"дыры в середине и в хвосте", reflect.TypeFor[struct {
			A bool
			B int64
			C bool
		}](), []Hole{{"A", 7}, {"C", 7}}},
		{"без дыр", reflect.TypeFor[struct {
			B int64
			A bool
			C bool
			D uint16
			E int32
		}](), nil},
		{"дыра перед int32", reflect.TypeFor[struct {
			A uint8
			B int32
		}](), []Hole{{"A", 3}}},
		{"паддинг внутри вложенной структуры не считается", reflect.TypeFor[struct {
			In inner
			Z  int64
		}](), nil},
		{"нулевое поле в конце даёт хвост", reflect.TypeFor[struct {
			A int64
			E struct{}
		}](), []Hole{{"E", 8}}},
		{"string после bool", reflect.TypeFor[struct {
			Ok   bool
			Name string
			N    int16
		}](), []Hole{{"Ok", 7}, {"N", 6}}},
		{"Event из первого задания", reflect.TypeFor[Event](),
			[]Hole{{"Active", 7}, {"Deleted", 7}, {"Kind", 7}, {"Retries", 2}}},
		{"не структура", reflect.TypeFor[int](), nil},
		{"пустая структура", reflect.TypeFor[struct{}](), nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Holes(tt.typ)
			if !slices.Equal(got, tt.want) {
				t.Errorf("Holes(%v) = %v, want %v", tt.typ, got, tt.want)
			}
		})
	}
}
