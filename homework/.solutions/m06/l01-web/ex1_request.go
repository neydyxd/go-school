package l01web

import (
	"errors"
	"fmt"
	"net/textproto"
	"strconv"
	"strings"
)

// ErrMalformed: запрос нарушает формат HTTP/1.x. Все ошибки ParseRequest
// оборачивают её, чтобы вызывающий код проверял errors.Is(err, ErrMalformed).
var ErrMalformed = errors.New("malformed request")

// Request: разобранный HTTP/1.x-запрос.
type Request struct {
	Method string              // "GET", "POST"...
	Target string              // как в стартовой строке: "/path?q=1"
	Proto  string              // "HTTP/1.1" или "HTTP/1.0"
	Header map[string][]string // ключи в каноническом виде: "Content-Type"
	Body   []byte
}

// ParseRequest разбирает сырой запрос raw, как его видит сервер в TCP-потоке.
//
// Формат: стартовая строка, заголовки, пустая строка, тело. Строки
// разделяются "\r\n", секция заголовков кончается "\r\n\r\n"; нет такой
// последовательности, значит запрос не дочитан: ошибка.
//
//   - Стартовая строка: ровно три части через один пробел: метод, цель, протокол.
//     Метод только из заглавных латинских букв A-Z. Цель начинается с "/".
//     Протокол "HTTP/1.1" или "HTTP/1.0".
//   - Заголовок: "Имя: значение". Имя непустое, без пробелов и табов
//     (пробел перед двоеточием запрещён RFC 9112: это вектор request smuggling).
//     Значение обрезается по краям от пробелов и табов. Имена приводятся
//     к каноническому виду (textproto.CanonicalMIMEHeaderKey), повторы
//     дописываются в срез в порядке появления.
//   - Host обязан встретиться ровно один раз для HTTP/1.1 (для 1.0 не обязателен,
//     но больше одного тоже ошибка).
//   - Transfer-Encoding не поддерживаем: ошибка.
//   - Content-Length: неотрицательное десятичное число; несколько разных значений
//     это ошибка. Тело: ровно столько БАЙТОВ после пустой строки. Байтов меньше:
//     ошибка; больше: лишнее не трогаем (это может быть следующий запрос).
//     Нет Content-Length: тело пустое (nil или пустой срез).
//   - Header никогда не nil: без заголовков это пустая мапа.
//
// Все ошибки оборачивают ErrMalformed.
func ParseRequest(raw string) (Request, error) {
	head, rest, ok := strings.Cut(raw, "\r\n\r\n")
	if !ok {
		return Request{}, fmt.Errorf("%w: нет конца заголовков", ErrMalformed)
	}
	lines := strings.Split(head, "\r\n")

	parts := strings.Split(lines[0], " ")
	if len(parts) != 3 {
		return Request{}, fmt.Errorf("%w: стартовая строка %q", ErrMalformed, lines[0])
	}
	req := Request{Method: parts[0], Target: parts[1], Proto: parts[2], Header: map[string][]string{}}
	if req.Method == "" || strings.Trim(req.Method, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
		return Request{}, fmt.Errorf("%w: метод %q", ErrMalformed, req.Method)
	}
	if !strings.HasPrefix(req.Target, "/") {
		return Request{}, fmt.Errorf("%w: цель %q", ErrMalformed, req.Target)
	}
	if req.Proto != "HTTP/1.1" && req.Proto != "HTTP/1.0" {
		return Request{}, fmt.Errorf("%w: протокол %q", ErrMalformed, req.Proto)
	}

	for _, line := range lines[1:] {
		name, value, ok := strings.Cut(line, ":")
		if !ok || name == "" || strings.ContainsAny(name, " \t") {
			return Request{}, fmt.Errorf("%w: заголовок %q", ErrMalformed, line)
		}
		key := textproto.CanonicalMIMEHeaderKey(name)
		req.Header[key] = append(req.Header[key], strings.Trim(value, " \t"))
	}

	if n := len(req.Header["Host"]); n > 1 || (n == 0 && req.Proto == "HTTP/1.1") {
		return Request{}, fmt.Errorf("%w: Host встречается %d раз", ErrMalformed, n)
	}
	if _, ok := req.Header["Transfer-Encoding"]; ok {
		return Request{}, fmt.Errorf("%w: Transfer-Encoding не поддерживается", ErrMalformed)
	}

	cls := req.Header["Content-Length"]
	if len(cls) == 0 {
		return req, nil
	}
	for _, v := range cls[1:] {
		if v != cls[0] {
			return Request{}, fmt.Errorf("%w: разные Content-Length", ErrMalformed)
		}
	}
	n, err := strconv.Atoi(cls[0])
	if err != nil || n < 0 || strings.HasPrefix(cls[0], "+") {
		return Request{}, fmt.Errorf("%w: Content-Length %q", ErrMalformed, cls[0])
	}
	if len(rest) < n {
		return Request{}, fmt.Errorf("%w: тело %d байт из %d", ErrMalformed, len(rest), n)
	}
	req.Body = []byte(rest[:n])
	return req, nil
}
