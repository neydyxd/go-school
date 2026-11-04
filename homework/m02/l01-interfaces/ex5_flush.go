package l01interfaces

import "io"

// Flusher: необязательная возможность writer-а сбросить буфер.
type Flusher interface {
	Flush() error
}

// FlushIfPossible сбрасывает буфер w, если w умеет это делать
// (реализует Flusher), и возвращает ошибку Flush.
// Если не умеет, ничего не делает и возвращает nil. w == nil: тоже nil.
func FlushIfPossible(w io.Writer) error {
	panic("TODO: реализуй FlushIfPossible")
}
