package l01memory

import "strings"

// Header это заголовок HTTP-запроса: имя и значение.
type Header struct {
	Name  string
	Value string
}

// ParseHeader: strings.Cut и strings.TrimSpace возвращают подстроки исходной
// строки (тот же массив байт, другие указатель и длина), поэтому ничего не
// аллоцируют. strings.Split или SplitN создали бы слайс в куче, а
// strings.ToLower или конкатенация создали бы новую строку.
func ParseHeader(line string) (h Header, ok bool) {
	name, value, found := strings.Cut(line, ":")
	if !found {
		return Header{}, false
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return Header{}, false
	}
	return Header{Name: name, Value: strings.TrimSpace(value)}, true
}
