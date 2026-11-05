package l04ci

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// ErrNoStatements: в профиле нет ни одного блока кода.
var ErrNoStatements = errors.New("l04ci: в профиле нет операторов")

// block: один блок покрытия, ключ "файл:начало,конец".
type block struct {
	file    string
	stmts   int
	covered bool
}

// parseProfile читает профиль и сливает повторы одного блока.
func parseProfile(r io.Reader) (map[string]*block, error) {
	sc := bufio.NewScanner(r)
	if !sc.Scan() {
		if err := sc.Err(); err != nil {
			return nil, err
		}
		return nil, errors.New("l04ci: пустой профиль, нет строки mode")
	}
	switch strings.TrimSpace(sc.Text()) {
	case "mode: set", "mode: count", "mode: atomic":
	default:
		return nil, fmt.Errorf("l04ci: строка 1: ждали mode: set|count|atomic, получили %q", sc.Text())
	}

	blocks := map[string]*block{}
	for n := 2; sc.Scan(); n++ {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		// file.go:10.20,12.3 2 1  (имя файла может содержать пробелы и двоеточия,
		// поэтому режем с конца)
		fields := strings.Fields(line)
		if len(fields) < 3 {
			return nil, fmt.Errorf("l04ci: строка %d: %q: ждали 3 поля", n, line)
		}
		pos := strings.Join(fields[:len(fields)-2], " ")
		stmts, err1 := strconv.Atoi(fields[len(fields)-2])
		count, err2 := strconv.Atoi(fields[len(fields)-1])
		colon := strings.LastIndex(pos, ":")
		if err1 != nil || err2 != nil || stmts < 0 || count < 0 || colon <= 0 {
			return nil, fmt.Errorf("l04ci: строка %d: %q: неверный формат", n, line)
		}
		b, ok := blocks[pos]
		if !ok {
			b = &block{file: pos[:colon], stmts: stmts}
			blocks[pos] = b
		}
		b.covered = b.covered || count > 0
	}
	return blocks, sc.Err()
}

// Coverage считает процент покрытия операторов по профилю из
// go test -coverprofile (тот же, что выводит go tool cover -func в total).
//
// Формат: первая строка "mode: set", "mode: count" или "mode: atomic", дальше
// строки вида "путь/файл.go:10.20,12.3 2 1": блок кода, число операторов в
// нём, сколько раз блок выполнился. Пустые строки пропускаются.
//
// Один и тот же блок может встретиться несколько раз (go test -coverpkg с
// несколькими пакетами пишет его из каждого тестового бинарника): операторы
// блока считаются один раз, а покрытым он считается, если хоть в одной
// строке счётчик больше нуля.
//
// Результат: 100 * покрытые / все операторы. Нет операторов: ErrNoStatements.
// Нет строки mode или битая строка: ошибка с номером строки ("строка 3").
func Coverage(r io.Reader) (float64, error) {
	blocks, err := parseProfile(r)
	if err != nil {
		return 0, err
	}
	var total, covered int
	for _, b := range blocks {
		total += b.stmts
		if b.covered {
			covered += b.stmts
		}
	}
	if total == 0 {
		return 0, ErrNoStatements
	}
	return 100 * float64(covered) / float64(total), nil
}
