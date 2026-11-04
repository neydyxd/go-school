package l06stdlib

import "io"

type repeatReader struct {
	s    string
	left int // сколько повторов ещё не отдано целиком
	off  int // смещение внутри текущего повтора
}

func NewRepeatReader(s string, n int) io.Reader {
	if s == "" || n < 0 {
		n = 0
	}
	return &repeatReader{s: s, left: n}
}

func (r *repeatReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.left == 0 {
		return 0, io.EOF
	}
	written := 0
	for written < len(p) && r.left > 0 {
		c := copy(p[written:], r.s[r.off:])
		written += c
		r.off += c
		if r.off == len(r.s) {
			r.off = 0
			r.left--
		}
	}
	return written, nil
}
