package l02controlflow

import "strconv"

// FizzBuzz возвращает строки для чисел от 1 до n включительно:
// кратное 15 даёт "FizzBuzz", кратное 3 даёт "Fizz", кратное 5 даёт "Buzz",
// остальные числа записываются цифрами ("1", "2", ...).
// Для n <= 0 возвращает пустой слайс (длины 0).
// Подсказка: strconv.Itoa превращает число в строку.
func FizzBuzz(n int) []string {
	res := []string{}
	for i := 1; i <= n; i++ {
		switch {
		case i%15 == 0: // порядок case важен: 15 проверяем раньше 3 и 5
			res = append(res, "FizzBuzz")
		case i%3 == 0:
			res = append(res, "Fizz")
		case i%5 == 0:
			res = append(res, "Buzz")
		default:
			res = append(res, strconv.Itoa(i))
		}
	}
	return res
}
