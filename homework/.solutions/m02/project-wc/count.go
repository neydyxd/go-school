// Package projectwc: ядро утилит в духе wc и grep поверх io.Reader.
package projectwc

import (
	"bufio"
	"io"
	"unicode"
)

type Stats struct {
	Lines int
	Words int
	Runes int
	Bytes int
}

func Count(r io.Reader) (Stats, error) {
	var st Stats
	br := bufio.NewReader(r)
	inWord := false
	for {
		ru, size, err := br.ReadRune()
		if err != nil {
			if err == io.EOF {
				return st, nil
			}
			return st, err
		}
		st.Runes++
		st.Bytes += size
		if ru == '\n' {
			st.Lines++
		}
		if unicode.IsSpace(ru) {
			inWord = false
		} else if !inWord {
			inWord = true
			st.Words++
		}
	}
}
