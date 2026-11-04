package l01types

// SafeAdd складывает a и b. Если сумма не помещается в int64
// (переполнение вверх или вниз), возвращает (0, false), иначе (сумма, true).
// Подсказка: сначала сложи, потом проверь знаки, или сравни с math.MaxInt64 - b.
func SafeAdd(a, b int64) (sum int64, ok bool) {
	panic("TODO: реализуй SafeAdd")
}
