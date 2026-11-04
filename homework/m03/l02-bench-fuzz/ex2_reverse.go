package l02benchfuzz

import "errors"

// ErrInvalidUTF8 возвращается, если на вход пришла невалидная UTF-8 строка.
var ErrInvalidUTF8 = errors.New("invalid UTF-8")

// Reverse разворачивает строку по рунам: Reverse("Привет") = "тевирП".
// Невалидная UTF-8 строка: "", ErrInvalidUTF8 (развернуть её по рунам
// нельзя без потери данных). Пустая строка: "", nil.
//
// Тест FuzzReverse проверяет свойства на семенах при обычном go test,
// а с флагом -fuzz ищет контрпримеры сам.
func Reverse(s string) (string, error) {
	panic("TODO: реализуй Reverse")
}
