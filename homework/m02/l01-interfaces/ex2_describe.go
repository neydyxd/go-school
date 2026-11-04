package l01interfaces

// Describe описывает значение по его динамическому типу. Используй type switch.
// Правила проверяются сверху вниз, срабатывает первое подходящее:
//
//	nil                  -> "nil"
//	error                -> "ошибка: <текст ошибки>"
//	fmt.Stringer         -> "stringer: <результат String()>"
//	int                  -> "int: <число>"
//	string               -> "строка из N рун" (N: число рун, не байт)
//	[]int                -> "[]int длины N"
//	bool                 -> "bool: true" или "bool: false"
//	всё остальное        -> "неизвестный тип <%T>", например "неизвестный тип float64"
//
// Значение, которое одновременно error и fmt.Stringer, описывается как ошибка.
func Describe(v any) string {
	panic("TODO: реализуй Describe")
}
