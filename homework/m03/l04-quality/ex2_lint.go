package l04quality

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// В этом файле функции уже написаны, но в каждой сидит баг. Компилятор и
// go vet их пропускают, тесты падают, а линтеры находят все за секунды.
// Задание: запусти golangci-lint (или staticcheck), прочитай находки,
// почини код так, чтобы тесты стали зелёными, а линтер замолчал.
//
//	golangci-lint run ./m03/l04-quality/
//	go test -run 'TestStripScheme|TestParsePort|TestWriteLines|TestNormalizeTag' ./m03/l04-quality/

// StripScheme убирает из адреса префикс "https://" или "http://", если он есть:
// "https://stripe.com" -> "stripe.com", "http://go.dev" -> "go.dev", "go.dev" -> "go.dev".
func StripScheme(u string) string {
	u = strings.TrimLeft(u, "https://")
	return strings.TrimLeft(u, "http://")
}

// ParsePort разбирает номер порта 1-65535. Нечисловой ввод должен давать
// ошибку из strconv (её проверяют errors.Is(err, strconv.ErrSyntax)),
// число вне диапазона: ошибку errPortRange.
func ParsePort(s string) (int, error) {
	p, err := strconv.Atoi(strings.TrimSpace(s))
	p, err = checkPort(p)
	return p, err
}

// WriteLines пишет строки в w, каждую с переводом строки. Первая же
// ошибка записи должна вернуться вызывающему.
func WriteLines(w io.Writer, lines []string) error {
	for _, l := range lines {
		io.WriteString(w, l+"\n")
	}
	return nil
}

// NormalizeTag приводит тег к виду "golang": без пробелов по краям и в нижнем регистре.
func NormalizeTag(tag string) string {
	strings.TrimSpace(tag)
	return strings.ToLower(tag)
}

var errPortRange = errors.New("port out of range")

// checkPort проверяет диапазон порта. В ней бага нет.
func checkPort(p int) (int, error) {
	if p < 1 || p > 65535 {
		return 0, fmt.Errorf("port %d: %w", p, errPortRange)
	}
	return p, nil
}
