package projectwc

import (
	"bufio"
	"io"
	"strings"
)

type Match struct {
	Line int
	Text string
}

type Options struct {
	IgnoreCase bool
	Invert     bool
}

func Grep(r io.Reader, pattern string, opt Options) ([]Match, error) {
	var out []Match
	err := GrepFunc(r, pattern, opt, func(m Match) error {
		out = append(out, m)
		return nil
	})
	return out, err
}

func GrepFunc(r io.Reader, pattern string, opt Options, fn func(Match) error) error {
	if opt.IgnoreCase {
		pattern = strings.ToLower(pattern)
	}
	br := bufio.NewReader(r)
	for n := 1; ; n++ {
		// ReadString не ограничен размером буфера, в отличие от bufio.Scanner.
		line, err := br.ReadString('\n')
		if line != "" {
			text := strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			hay := text
			if opt.IgnoreCase {
				hay = strings.ToLower(hay)
			}
			if strings.Contains(hay, pattern) != opt.Invert {
				if ferr := fn(Match{Line: n, Text: text}); ferr != nil {
					return ferr
				}
			}
		}
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
	}
}
