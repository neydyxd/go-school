package l06strings

import "strings"

// CSVLine собирает строку CSV из полей через запятую.
//
// Поле, в котором есть запятая, двойная кавычка или перевод строки,
// заключается в двойные кавычки, а кавычки внутри него удваиваются:
// []string{"a", "b,c", `say "hi"`} -> `a,"b,c","say ""hi"""`.
// Пустой список полей: пустая строка.
//
// Собирай через strings.Builder. Звёздочка: заранее посчитай итоговую длину
// и вызови Grow, чтобы вся строка строилась за одну аллокацию.
func CSVLine(fields []string) string {
	if len(fields) == 0 {
		return ""
	}
	size := len(fields) - 1 // запятые
	for _, f := range fields {
		size += len(f)
		if needsQuotes(f) {
			size += 2 + strings.Count(f, `"`)
		}
	}
	var b strings.Builder
	b.Grow(size)
	for i, f := range fields {
		if i > 0 {
			b.WriteByte(',')
		}
		if !needsQuotes(f) {
			b.WriteString(f)
			continue
		}
		b.WriteByte('"')
		for _, r := range f {
			if r == '"' {
				b.WriteByte('"')
			}
			b.WriteRune(r)
		}
		b.WriteByte('"')
	}
	return b.String()
}

func needsQuotes(f string) bool {
	return strings.ContainsAny(f, ",\"\n")
}
