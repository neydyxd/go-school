package l06stdlib

import (
	"errors"
	"io"
)

var (
	// ErrBadLine: строка не в формате key=value или пустой ключ.
	ErrBadLine = errors.New("bad line")
	// ErrDuplicate: ключ встретился второй раз.
	ErrDuplicate = errors.New("duplicate key")
)

// ParseConfig читает конфиг из r построчно и возвращает мапу ключ -> целое число.
//
// Формат:
//   - пробелы по краям строки, ключа и значения игнорируются;
//   - пустые строки и строки, начинающиеся с "#", пропускаются;
//   - остальные строки: key=value (делим по ПЕРВОМУ "="), value это целое число;
//   - строки могут быть очень длинными (сотни килобайт), это не ошибка.
//
// Ошибки (вернуть nil-мапу и ошибку с номером строки, нумерация с 1):
//   - нет "=" или пустой ключ: текст "line N: ...", errors.Is(err, ErrBadLine);
//   - значение не число: "line N: ...", и через errors.Is видна ошибка strconv
//     (strconv.ErrSyntax или strconv.ErrRange);
//   - повтор ключа: "line N: ...", errors.Is(err, ErrDuplicate);
//   - ошибка чтения из r: вернуть её, обёрнутую (errors.Is должен её находить).
func ParseConfig(r io.Reader) (map[string]int, error) {
	panic("TODO: реализуй ParseConfig")
}
