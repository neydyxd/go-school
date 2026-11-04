package l03functions

// Compose склеивает функции в конвейер слева направо:
// Compose(f, g, h)(x) == h(g(f(x))).
// Без аргументов возвращает тождественную функцию: Compose()(x) == x.
//
// Важно: если вызвать Compose(fs...) со слайсом, а потом поменять элементы fs,
// уже собранный конвейер меняться не должен.
func Compose(fs ...func(int) int) func(int) int {
	panic("TODO: реализуй Compose")
}
