package l06stdlib

import "io"

// NewRepeatReader возвращает io.Reader, который отдаёт строку s ровно n раз подряд,
// а затем io.EOF. n <= 0 или пустая s: сразу io.EOF.
//
// Не склеивай всё в одну большую строку заранее: n может быть огромным.
// Храни позицию и заполняй буфер p в Read кусками. Соблюдай контракт io.Reader:
// возвращай 0 <= n <= len(p); при len(p) == 0 верни 0, nil.
func NewRepeatReader(s string, n int) io.Reader {
	panic("TODO: реализуй NewRepeatReader")
}
