package l05packages

import (
	"strconv"
	"strings"
)

func allDigits(s string) bool {
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

func SplitPathVersion(path string) (prefix string, major int, ok bool) {
	if path == "" {
		return "", 0, false
	}
	i := strings.LastIndex(path, "/")
	num, isV := strings.CutPrefix(path[i+1:], "v")
	if i < 0 || !isV || !allDigits(num) {
		return path, 1, true
	}
	if num[0] == '0' || i == 0 {
		return "", 0, false
	}
	n, err := strconv.Atoi(num)
	if err != nil || n < 2 {
		return "", 0, false
	}
	return path[:i], n, true
}

func CompatibleVersion(path, version string) bool {
	_, pm, ok := SplitPathVersion(path)
	if !ok {
		return false
	}
	rest, isV := strings.CutPrefix(version, "v")
	num, _, hasDot := strings.Cut(rest, ".")
	if !isV || !hasDot || !allDigits(num) {
		return false
	}
	vm, err := strconv.Atoi(num)
	if err != nil {
		return false
	}
	if pm == 1 {
		return vm <= 1
	}
	return vm == pm
}
