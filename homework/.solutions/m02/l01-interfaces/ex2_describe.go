package l01interfaces

import (
	"fmt"
	"strconv"
	"unicode/utf8"
)

func Describe(v any) string {
	switch x := v.(type) {
	case nil:
		return "nil"
	case error:
		return "ошибка: " + x.Error()
	case fmt.Stringer:
		return "stringer: " + x.String()
	case int:
		return "int: " + strconv.Itoa(x)
	case string:
		return fmt.Sprintf("строка из %d рун", utf8.RuneCountInString(x))
	case []int:
		return fmt.Sprintf("[]int длины %d", len(x))
	case bool:
		return "bool: " + strconv.FormatBool(x)
	default:
		return fmt.Sprintf("неизвестный тип %T", x)
	}
}
