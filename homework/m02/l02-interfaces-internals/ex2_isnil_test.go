package l02interfacesinternals

import (
	"io"
	"testing"
)

func TestIsNil(t *testing.T) {
	var (
		np   *int
		ns   []int
		nm   map[string]int
		nc   chan int
		nf   func()
		nve  *ValidationError
		x    = 5
		nerr error
	)
	tests := []struct {
		name string
		in   any
		want bool
	}{
		{"голый nil", nil, true},
		{"nil-интерфейс error", nerr, true},
		{"nil-указатель", np, true},
		{"nil-слайс", ns, true},
		{"nil-мапа", nm, true},
		{"nil-канал", nc, true},
		{"nil-функция", nf, true},
		{"nil-указатель на свой тип ошибки", nve, true},
		{"nil-указатель в io.Reader", io.Reader((*badReader)(nil)), true},
		{"ноль", 0, false},
		{"пустая строка", "", false},
		{"пустая структура", struct{}{}, false},
		{"массив", [2]int{}, false},
		{"живой указатель", &x, false},
		{"пустой, но не nil слайс", []int{}, false},
		{"пустая мапа", map[string]int{}, false},
		{"false", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsNil(tt.in); got != tt.want {
				t.Errorf("IsNil(%#v) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

type badReader struct{}

func (*badReader) Read([]byte) (int, error) { return 0, io.EOF }
