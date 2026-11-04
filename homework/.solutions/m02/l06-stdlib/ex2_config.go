package l06stdlib

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	ErrBadLine   = errors.New("bad line")
	ErrDuplicate = errors.New("duplicate key")
)

func ParseConfig(r io.Reader) (map[string]int, error) {
	out := map[string]int{}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 16<<20)
	n := 0
	for sc.Scan() {
		n++
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		k, v = strings.TrimSpace(k), strings.TrimSpace(v)
		if !ok || k == "" {
			return nil, fmt.Errorf("line %d: %q: %w", n, line, ErrBadLine)
		}
		num, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", n, err)
		}
		if _, dup := out[k]; dup {
			return nil, fmt.Errorf("line %d: %q: %w", n, k, ErrDuplicate)
		}
		out[k] = num
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	return out, nil
}
