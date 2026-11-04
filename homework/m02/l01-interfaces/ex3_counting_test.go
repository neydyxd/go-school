package l01interfaces

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
)

var _ io.Writer = (*CountingWriter)(nil)

// halfWriter записывает половину данных и возвращает ошибку.
type halfWriter struct{ buf bytes.Buffer }

var errDiskFull = errors.New("диск полон")

func (h *halfWriter) Write(p []byte) (int, error) {
	n := len(p) / 2
	h.buf.Write(p[:n])
	return n, errDiskFull
}

func TestCountingWriter(t *testing.T) {
	t.Run("fmt.Fprintf с кириллицей", func(t *testing.T) {
		var buf bytes.Buffer
		cw := NewCountingWriter(&buf)
		fmt.Fprintf(cw, "привет %d", 42)
		if buf.String() != "привет 42" {
			t.Errorf("в буфер записано %q, want %q", buf.String(), "привет 42")
		}
		if got := cw.Count(); got != 15 {
			t.Errorf("Count() = %d, want 15 (байты, а не руны)", got)
		}
	})
	t.Run("несколько записей копятся", func(t *testing.T) {
		cw := NewCountingWriter(io.Discard)
		io.WriteString(cw, "abc")
		io.WriteString(cw, "")
		io.Copy(cw, strings.NewReader(strings.Repeat("x", 10000)))
		if got := cw.Count(); got != 10003 {
			t.Errorf("Count() = %d, want 10003", got)
		}
	})
	t.Run("частичная запись с ошибкой", func(t *testing.T) {
		hw := &halfWriter{}
		cw := NewCountingWriter(hw)
		n, err := cw.Write([]byte("12345678"))
		if n != 4 || !errors.Is(err, errDiskFull) {
			t.Errorf("Write(8 байт) = %d, %v, want 4, %v", n, err, errDiskFull)
		}
		if got := cw.Count(); got != 4 {
			t.Errorf("Count() = %d, want 4", got)
		}
	})
	t.Run("новый счётчик с нуля", func(t *testing.T) {
		if got := NewCountingWriter(io.Discard).Count(); got != 0 {
			t.Errorf("Count() = %d, want 0", got)
		}
	})
}
