package l01interfaces

import (
	"bufio"
	"bytes"
	"errors"
	"testing"
)

type failFlusher struct{ bytes.Buffer }

var errFlush = errors.New("сеть отвалилась")

func (f *failFlusher) Flush() error { return errFlush }

func TestFlushIfPossible(t *testing.T) {
	t.Run("bufio.Writer сбрасывается", func(t *testing.T) {
		var dst bytes.Buffer
		bw := bufio.NewWriter(&dst)
		bw.WriteString("данные")
		if dst.Len() != 0 {
			t.Fatalf("до Flush данные уже в dst: %q", dst.String())
		}
		if err := FlushIfPossible(bw); err != nil {
			t.Fatalf("FlushIfPossible(bufio.Writer) = %v, want nil", err)
		}
		if dst.String() != "данные" {
			t.Errorf("после FlushIfPossible dst = %q, want %q", dst.String(), "данные")
		}
	})
	t.Run("bytes.Buffer не флашер", func(t *testing.T) {
		if err := FlushIfPossible(&bytes.Buffer{}); err != nil {
			t.Errorf("FlushIfPossible(bytes.Buffer) = %v, want nil", err)
		}
	})
	t.Run("ошибка Flush пробрасывается", func(t *testing.T) {
		if err := FlushIfPossible(&failFlusher{}); !errors.Is(err, errFlush) {
			t.Errorf("FlushIfPossible(failFlusher) = %v, want %v", err, errFlush)
		}
	})
	t.Run("nil writer", func(t *testing.T) {
		if err := FlushIfPossible(nil); err != nil {
			t.Errorf("FlushIfPossible(nil) = %v, want nil", err)
		}
	})
}
