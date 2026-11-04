package l06stdlib

import (
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestRepeatReader(t *testing.T) {
	tests := []struct {
		name string
		s    string
		n    int
	}{
		{"три раза", "ab", 3},
		{"один раз", "привет", 1},
		{"ноль раз", "x", 0},
		{"отрицательное n", "x", -5},
		{"пустая строка", "", 10},
		{"длинный результат", "0123456789", 5000},
		{"кириллица много раз", "ёж,", 777},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := strings.Repeat(tt.s, max(tt.n, 0))
			got, err := io.ReadAll(NewRepeatReader(tt.s, tt.n))
			if err != nil || string(got) != want {
				t.Fatalf("ReadAll(NewRepeatReader(%q, %d)) = %d байт, %v, want %d байт, nil", tt.s, tt.n, len(got), err, len(want))
			}
			// iotest.TestReader проверяет контракт: чтение мелкими кусками,
			// поведение после EOF, ReadAt/Seek если они есть.
			if err := iotest.TestReader(NewRepeatReader(tt.s, tt.n), []byte(want)); err != nil {
				t.Errorf("iotest.TestReader(NewRepeatReader(%q, %d)): %v", tt.s, tt.n, err)
			}
		})
	}
	t.Run("по одному байту", func(t *testing.T) {
		got, err := io.ReadAll(iotest.OneByteReader(NewRepeatReader("abc", 4)))
		if err != nil || string(got) != "abcabcabcabc" {
			t.Errorf("OneByteReader: %q, %v, want %q", got, err, "abcabcabcabc")
		}
	})
	t.Run("пустой буфер", func(t *testing.T) {
		n, err := NewRepeatReader("abc", 1).Read(nil)
		if n != 0 || err != nil {
			t.Errorf("Read(nil) = %d, %v, want 0, nil", n, err)
		}
	})
	t.Run("огромное n без выделения памяти заранее", func(t *testing.T) {
		r := NewRepeatReader("go", 1<<40)
		buf := make([]byte, 5)
		n, err := io.ReadFull(r, buf)
		if n != 5 || err != nil || string(buf) != "gogog" {
			t.Errorf("первые 5 байт = %q, %v, want %q", buf[:n], err, "gogog")
		}
	})
}
