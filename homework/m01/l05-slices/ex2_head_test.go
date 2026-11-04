package l05slices

import (
	"bytes"
	"testing"
)

func TestHead(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		n    int
		want []byte
	}{
		{"обычный случай", []byte("hello, world"), 5, []byte("hello")},
		{"n равен длине", []byte("abc"), 3, []byte("abc")},
		{"n больше длины", []byte("abc"), 10, []byte("abc")},
		{"n ноль", []byte("abc"), 0, []byte{}},
		{"n отрицательный", []byte("abc"), -1, []byte{}},
		{"nil на входе", nil, 3, []byte{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Head(tt.data, tt.n)
			if !bytes.Equal(got, tt.want) || len(got) != len(tt.want) {
				t.Errorf("Head(%q, %d) = %q, want %q", tt.data, tt.n, got, tt.want)
			}
		})
	}
}

func TestHeadIsCopy(t *testing.T) {
	big := make([]byte, 1<<20)
	copy(big, "header")
	got := Head(big, 6)
	if string(got) != "header" {
		t.Fatalf("Head(big, 6) = %q, want %q", got, "header")
	}
	got[0] = 'H'
	if big[0] != 'h' {
		t.Errorf("запись в результат Head изменила исходный буфер: результат смотрит в тот же массив")
	}
	if cap(got) > 1024 {
		t.Errorf("cap(Head(big, 6)) = %d: результат держит огромный массив, нужна копия", cap(got))
	}
}
