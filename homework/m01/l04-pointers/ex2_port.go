package l04pointers

// ParsePort разбирает номер порта из строки.
// Корректный порт: целое число от 1 до 65535 без пробелов и знака "+".
// Для корректного порта возвращает указатель на число, иначе nil.
// nil здесь означает "значения нет", как null в JSON.
// Подсказка: strconv.Atoi; "+80" он считает числом, это нужно отсечь отдельно.
func ParsePort(s string) *int {
	panic("TODO: реализуй ParsePort")
}

// PortOr возвращает *p, а если p == nil, то def.
func PortOr(p *int, def int) int {
	panic("TODO: реализуй PortOr")
}
