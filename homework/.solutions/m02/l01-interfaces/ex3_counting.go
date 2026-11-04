package l01interfaces

import "io"

type CountingWriter struct {
	w io.Writer
	n int64
}

func NewCountingWriter(w io.Writer) *CountingWriter {
	return &CountingWriter{w: w}
}

func (c *CountingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.n += int64(n)
	return n, err
}

func (c *CountingWriter) Count() int64 { return c.n }
