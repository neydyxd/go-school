package l04quality

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// StripScheme убирает из адреса префикс "https://" или "http://", если он есть.
// TrimLeft принимает набор символов (cutset), а не префикс: нужен TrimPrefix.
func StripScheme(u string) string {
	if s, ok := strings.CutPrefix(u, "https://"); ok {
		return s
	}
	return strings.TrimPrefix(u, "http://")
}

// ParsePort разбирает номер порта 1-65535.
func ParsePort(s string) (int, error) {
	p, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0, err
	}
	return checkPort(p)
}

// WriteLines пишет строки в w, каждую с переводом строки.
func WriteLines(w io.Writer, lines []string) error {
	for _, l := range lines {
		if _, err := io.WriteString(w, l+"\n"); err != nil {
			return err
		}
	}
	return nil
}

// NormalizeTag приводит тег к виду "golang".
func NormalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

var errPortRange = errors.New("port out of range")

// checkPort проверяет диапазон порта. В ней бага нет.
func checkPort(p int) (int, error) {
	if p < 1 || p > 65535 {
		return 0, fmt.Errorf("port %d: %w", p, errPortRange)
	}
	return p, nil
}
