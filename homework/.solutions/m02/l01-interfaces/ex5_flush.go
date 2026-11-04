package l01interfaces

import "io"

type Flusher interface {
	Flush() error
}

func FlushIfPossible(w io.Writer) error {
	if f, ok := w.(Flusher); ok {
		return f.Flush()
	}
	return nil
}
