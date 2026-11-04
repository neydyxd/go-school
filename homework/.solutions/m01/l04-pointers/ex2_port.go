package l04pointers

import "strconv"

// ParsePort разбирает номер порта из строки.
// Корректный порт: целое число от 1 до 65535 без пробелов и знака "+".
// Для корректного порта возвращает указатель на число, иначе nil.
// nil здесь означает "значения нет", как null в JSON.
// Подсказка: strconv.Atoi; "+80" он считает числом, это нужно отсечь отдельно.
func ParsePort(s string) *int {
	if s == "" || s[0] == '+' {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 || n > 65535 {
		return nil
	}
	return &n // n уйдёт в кучу, и это безопасно: в Go нет висячих указателей
}

// PortOr возвращает *p, а если p == nil, то def.
func PortOr(p *int, def int) int {
	if p == nil {
		return def
	}
	return *p
}
