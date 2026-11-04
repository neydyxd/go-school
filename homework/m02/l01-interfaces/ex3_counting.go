package l01interfaces

import "io"

// CountingWriter оборачивает другой io.Writer и считает, сколько байт
// через него реально записано. Сам он тоже io.Writer, поэтому его можно
// передать в fmt.Fprintf, io.Copy и куда угодно ещё.
type CountingWriter struct {
	w io.Writer
	n int64
}

// NewCountingWriter принимает интерфейс и возвращает конкретный тип.
func NewCountingWriter(w io.Writer) *CountingWriter {
	panic("TODO: реализуй NewCountingWriter")
}

// Write пишет p в обёрнутый writer и прибавляет к счётчику ровно столько
// байт, сколько тот записал (даже если вернул ошибку). n и err верни как есть.
func (c *CountingWriter) Write(p []byte) (int, error) {
	panic("TODO: реализуй Write")
}

// Count возвращает число успешно записанных байт.
func (c *CountingWriter) Count() int64 {
	panic("TODO: реализуй Count")
}
