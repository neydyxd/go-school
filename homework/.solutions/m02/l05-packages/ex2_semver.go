package l05packages

import (
	"cmp"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var ErrInvalidVersion = errors.New("invalid version")

type version struct {
	nums [3]int
	pre  []string
}

func isNum(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func parseVersion(s string) (version, error) {
	var v version
	bad := fmt.Errorf("%q: %w", s, ErrInvalidVersion)
	rest, ok := strings.CutPrefix(s, "v")
	if !ok {
		return v, bad
	}
	core, pre, hasPre := strings.Cut(rest, "-")
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return v, bad
	}
	for i, p := range parts {
		if !isNum(p) || (len(p) > 1 && p[0] == '0') {
			return v, bad
		}
		n, err := strconv.Atoi(p)
		if err != nil {
			return v, bad
		}
		v.nums[i] = n
	}
	if hasPre {
		v.pre = strings.Split(pre, ".")
		for _, id := range v.pre {
			if id == "" || (isNum(id) && len(id) > 1 && id[0] == '0') {
				return v, bad
			}
		}
	}
	return v, nil
}

func comparePre(a, b []string) int {
	if len(a) == 0 || len(b) == 0 {
		// Отсутствие пререлиза старше любого пререлиза.
		return cmp.Compare(len(b), len(a))
	}
	for i := range min(len(a), len(b)) {
		x, y := a[i], b[i]
		xn, yn := isNum(x), isNum(y)
		var c int
		switch {
		case xn && yn:
			c = cmp.Compare(len(x), len(y)) // без ведущих нулей длиннее значит больше
			if c == 0 {
				c = strings.Compare(x, y)
			}
		case xn:
			c = -1
		case yn:
			c = 1
		default:
			c = strings.Compare(x, y)
		}
		if c != 0 {
			return c
		}
	}
	return cmp.Compare(len(a), len(b))
}

func CompareVersions(a, b string) (int, error) {
	va, err := parseVersion(a)
	if err != nil {
		return 0, err
	}
	vb, err := parseVersion(b)
	if err != nil {
		return 0, err
	}
	for i := range va.nums {
		if c := cmp.Compare(va.nums[i], vb.nums[i]); c != 0 {
			return c, nil
		}
	}
	return comparePre(va.pre, vb.pre), nil
}
